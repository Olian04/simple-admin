package upstream

import (
	"context"
	"net/http"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

type credentialKey struct{}

// WithCredential stores the passthrough credential (supplied by the browser via
// the credentials middleware) on the context.
func WithCredential(ctx context.Context, cred string) context.Context {
	return context.WithValue(ctx, credentialKey{}, cred)
}

// CredentialFromContext returns the passthrough credential, if any.
func CredentialFromContext(ctx context.Context) (string, bool) {
	cred, ok := ctx.Value(credentialKey{}).(string)
	return cred, ok && cred != ""
}

// applyAuth injects the context credential into the upstream request per the
// resource's auth scheme. none → no-op; apikey → header/query; bearer/basic/
// oauth → not implemented (MVP).
func applyAuth(ctx context.Context, req *http.Request, scheme resource.AuthScheme) {
	if scheme.Type != "apikey" {
		return
	}
	cred, ok := CredentialFromContext(ctx)
	if !ok {
		return
	}
	switch scheme.In {
	case "query":
		q := req.URL.Query()
		q.Set(scheme.Name, cred)
		req.URL.RawQuery = q.Encode()
	default: // header
		req.Header.Set(scheme.Name, cred)
	}
}
