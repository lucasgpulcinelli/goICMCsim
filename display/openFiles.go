package display

import (
	"github.com/sqweek/dialog"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// openFile is an alternative solution to problems encountered with the dialog.NewFileOpen() function.
// This function opens a file selection dialog and returns a fyne.URIReadCloser for the selected file.
// It uses the sqweek/dialog library to open the file selection dialog and the fyne library
// library to create a fyne.URIReadCloser from the selected file. 
func openFile(callback func(fyne.URIReadCloser, error)) {
	filename, err := dialog.File().Load()
	if err != nil {
		callback(nil, err)
		return
	}

	uri := storage.NewFileURI(filename)
	reader, err := storage.Reader(uri)
	callback(reader, err)
}