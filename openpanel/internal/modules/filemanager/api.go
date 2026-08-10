package filemanager

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires the /api/ equivalents of the file manager's routes
// onto mux, gated behind the "filemanager" feature flag. Several handlers
// already speak pure JSON with no session/flash dependency
// (handleFolders, handleDeleteFile, handleCopyItem, handleMoveItem,
// handleWgetFiles, handleWgetStatus) and are reused here unmodified; the
// listing/edit-file GET handlers already support "?output=json" and are
// invoked with that forced on via forceJSONOutput. Everything else
// (create/rename/permissions/upload/archive/save) needed a JSON-response
// variant, split across api_crud.go, api_content.go, and api_transfer.go.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "filemanager", "GET /api/files", func(w http.ResponseWriter, r *http.Request) {
		handleFiles(a, w, forceJSONOutput(r), "")
	})
	// "GET /api/files/{path_param...}" doesn't cover the bare "/api/files"
	// case (no trailing slash) - see filemanager.go's identical note on the
	// web routes this mirrors.
	apiregistry.Handle(mux, a, "filemanager", "GET /api/files/{path_param...}", func(w http.ResponseWriter, r *http.Request) {
		handleFiles(a, w, forceJSONOutput(r), r.PathValue("path_param"))
	})
	apiregistry.Handle(mux, a, "filemanager", "GET /api/folders", func(w http.ResponseWriter, r *http.Request) { handleFolders(a, w, r, "") })
	apiregistry.Handle(mux, a, "filemanager", "GET /api/folders/{path_param...}", func(w http.ResponseWriter, r *http.Request) {
		handleFolders(a, w, r, r.PathValue("path_param"))
	})

	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/new-file", func(w http.ResponseWriter, r *http.Request) { apiCreateFile(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/new-directory", func(w http.ResponseWriter, r *http.Request) { apiCreateFolder(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/rename", func(w http.ResponseWriter, r *http.Request) { apiRenameFile(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/permissions", func(w http.ResponseWriter, r *http.Request) { apiChangePermissions(a, w, r) })

	apiregistry.Handle(mux, a, "filemanager", "DELETE /api/file-manager/delete", func(w http.ResponseWriter, r *http.Request) { handleDeleteFile(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/copy", func(w http.ResponseWriter, r *http.Request) { handleCopyItem(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/move", func(w http.ResponseWriter, r *http.Request) { handleMoveItem(a, w, r) })

	apiregistry.Handle(mux, a, "filemanager", "GET /api/file-manager/edit-file/{file_path...}", func(w http.ResponseWriter, r *http.Request) {
		apiGetFileContent(a, w, r, r.PathValue("file_path"))
	})
	apiregistry.Handle(mux, a, "filemanager", "PUT /api/file-manager/edit-file/{file_path...}", func(w http.ResponseWriter, r *http.Request) {
		apiSaveFileContent(a, w, r, r.PathValue("file_path"))
	})
	apiregistry.Handle(mux, a, "filemanager", "GET /api/file-manager/download-file/{filename...}", func(w http.ResponseWriter, r *http.Request) {
		apiDownloadFile(a, w, r, r.PathValue("filename"))
	})
	apiregistry.Handle(mux, a, "filemanager", "GET /api/file-manager/view-file/{filename...}", func(w http.ResponseWriter, r *http.Request) {
		apiViewFile(a, w, r, r.PathValue("filename"))
	})

	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/upload", func(w http.ResponseWriter, r *http.Request) { apiUploadFiles(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/extract-archive", func(w http.ResponseWriter, r *http.Request) { apiExtractFiles(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/create-archive", func(w http.ResponseWriter, r *http.Request) { apiCompressFiles(a, w, r) })

	apiregistry.Handle(mux, a, "filemanager", "POST /api/file-manager/wget", func(w http.ResponseWriter, r *http.Request) { handleWgetFiles(a, w, r) })
	apiregistry.Handle(mux, a, "filemanager", "GET /api/file-manager/wget/status/{download_id}", func(w http.ResponseWriter, r *http.Request) { handleWgetStatus(a, w, r) })
}

// forceJSONOutput clones r with output=json set on the query string, so a
// handler that already branches on "?output=json" for its API-shaped
// response can be reused unmodified for a dedicated /api/ route.
func forceJSONOutput(r *http.Request) *http.Request {
	q := r.URL.Query()
	q.Set("output", "json")
	r2 := r.Clone(r.Context())
	r2.URL.RawQuery = q.Encode()
	return r2
}
