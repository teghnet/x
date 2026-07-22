package secrets

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/xakep666/gkpxc"

	"github.com/teghnet/x"
	"github.com/teghnet/x/parse/json"
)

func NewKeepassXCProvider(ctx context.Context, path string) *KeepassXCProvider {
	return &KeepassXCProvider{
		ctx:  ctx,
		path: path,
	}
}

// KeepassXCProvider reads secrets from a local KeepassXC instance via its
// "Secret Service Integration" setting (Application Settings → Secret
// Service Integration), which exposes the unlocked database over the same
// org.freedesktop.secrets D-Bus interface as any other Linux keyring. The
// master password never touches accd — the database just needs to be open
// when accd runs. Each secret is a KeepassXC entry titled to match key
// (e.g. "stripe.api_key"), grouped under Service.
type KeepassXCProvider struct {
	ctx  context.Context
	path string
}

func (p *KeepassXCProvider) Get(key string) (string, error) {
	u, err := url.Parse(key)
	if err != nil {
		return "", err
	}

	c, err := client(p.ctx, p.path)
	if err != nil {
		return "", err
	}
	defer x.Close(c)

	q := u.Query()
	var filters []entryFilter
	if q.Has("empty") {
		filters = append(filters, emptyLogin)
	}
	if q.Has("non-empty") {
		filters = append(filters, nonEmptyLogin)
	}
	if v := q.Get("login"); v != "" {
		filters = append(filters, login(v))
	}
	if v := q.Get("prefix"); v != "" {
		filters = append(filters, nameHasPrefix(v))
	}
	id := q.Get("uuid")
	if id != "" {
		filters = append(filters, uuid(id))
	}
	u.RawQuery = ""

	if q.Has("totp") {
		return totp(p.ctx, c, id)
	}

	e, err := entry(p.ctx, c, u.String(), filters...)
	if err != nil {
		return "", err
	}

	return e.Password, nil
}

type loginEntry gkpxc.LoginEntry

func (le loginEntry) String() string { return fmt.Sprintf("%s %s", le.Name, le.Login) }

type entryFilter func(loginEntry) bool

func emptyLogin(le loginEntry) bool    { return le.Login == "" }
func nonEmptyLogin(le loginEntry) bool { return le.Login != "" }
func uuid(val string) entryFilter      { return func(le loginEntry) bool { return le.UUID == val } }
func login(val string) entryFilter     { return func(le loginEntry) bool { return le.Login == val } }

func nameHasPrefix(prefix string) entryFilter {
	return func(le loginEntry) bool { return strings.HasPrefix(le.Name, prefix) }
}

func entry(ctx context.Context, c *gkpxc.Client, url string, filters ...entryFilter) (loginEntry, error) {
	ll, err := entries(ctx, c, url, filters...)
	if err != nil {
		return loginEntry{}, err
	}
	if len(ll) > 1 {
		return loginEntry{}, errors.New("too many results")
	}
	if len(ll) == 0 {
		return loginEntry{}, errors.New("not found")
	}
	return ll[0], nil
}

func entries(ctx context.Context, c *gkpxc.Client, url string, filters ...entryFilter) ([]loginEntry, error) {
	logins, err := c.GetLogins(ctx, gkpxc.GetLoginsRequest{
		URL: url,
	})
	if err != nil {
		return nil, err
	}
	var ll []loginEntry
	for _, l := range logins.Entries {
		if !filter(loginEntry(l), filters...) {
			continue
		}
		ll = append(ll, loginEntry(l))
	}
	return ll, nil
}

func totp(ctx context.Context, c *gkpxc.Client, id string) (string, error) {
	pass, err := c.GetTOTP(ctx, gkpxc.GetTOTPRequest{
		UUID: id,
	})
	if err != nil {
		return "", err
	}
	return pass.TOTP, nil
}

func filter(entry loginEntry, filters ...entryFilter) bool {
	for _, f := range filters {
		if !f(entry) {
			return false
		}
	}
	return true
}

func client(ctx context.Context, path string) (*gkpxc.Client, error) {
	socketPath := fmt.Sprintf("/run/user/%d/app/org.keepassxc.KeePassXC/%s", os.Getuid(), gkpxc.SocketName)
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the socket (%s): %v", socketPath, err)
	}
	c, err := gkpxc.NewClient(ctx, gkpxc.WithConn(conn))
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	_, err = c.GetDatabaseHash(ctx, true)
	if err != nil {
		return nil, err
	}

	err = associate(ctx, c, path)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func associate(ctx context.Context, client *gkpxc.Client, path string) error {
	if isFile(path) {
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer x.Close(r)

		c, err := json.Decode[gkpxc.AssociationCredentials](r)
		if err != nil {
			return err
		}
		client.SetAssociationCredentials(&c)

		return client.TestAssociate(ctx)
	}
	if err := client.Associate(ctx); err != nil {
		return err
	}

	c := client.AssociationCredentials()
	x.Infof("Association ID: %s", c.ID)

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer x.Close(w)

	return json.Write(w, c)
}

func isFile(elems ...string) bool {
	info, err := os.Stat(path.Join(elems...))
	return err == nil && !info.IsDir()
}
