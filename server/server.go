package server

// def MIME types
var StaticPostfix = map[string]string{
	".js":    "application/javascript; charset=utf-8",
	".css":   "text/css; charset=utf-8",
	".svg":   "image/svg+xml",
	".json":  "text/json",
	".ico":   "image/x-icon",
	".cur":   "image/x-icon",
	".png":   "image/png",
	".bmp":   "image/bmp",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".jfif":  "image/jpeg",
	".pjpeg": "image/jpeg",
	".pjp":   "image/jpeg",
	".gif":   "image/gif",
	".tif":   "image/tiff",
	".tiff":  "image/tiff",
	".webp":  "image/webp",
}
