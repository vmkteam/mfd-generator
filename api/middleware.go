package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/vmkteam/mfd-generator/mfd"

	genna "github.com/dizzyfool/genna/lib"
	"github.com/vmkteam/zenrpc/v2"
)

type Store struct {
	Genna          *genna.Genna
	CurrentFile    string
	CurrentProject *mfd.Project
}

func ProjectMiddleware(store *Store) zenrpc.MiddlewareFunc {
	return func(handler zenrpc.InvokeFunc) zenrpc.InvokeFunc {
		return func(ctx context.Context, method string, params json.RawMessage) zenrpc.Response {
			requestID := zenrpc.IDFromContext(ctx)
			namespace := zenrpc.NamespaceFromContext(ctx)

			if namespace == publicNS {
				return handler(ctx, method, params)
			}

			if namespace == projectNS && method == RPC.ProjectService.Open {
				return handler(ctx, method, params)
			}

			if store.CurrentProject != nil && store.Genna != nil {
				return handler(ctx, method, params)
			}

			return zenrpc.NewResponseError(requestID, http.StatusBadRequest, "project not opened", nil)
		}
	}
}
