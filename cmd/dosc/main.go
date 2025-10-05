package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"dosc/internal/bookmarks"
	"dosc/internal/logger"
	"dosc/internal/s3client"
)

var masterPassword string
var bookmarkList *widget.List
var currentBookmarks []bookmarks.Bookmark
var activeBookmark *bookmarks.Bookmark
var s3Client *s3client.S3Client

var remoteFiles *widget.List
var remoteFileItems []string
var localFiles *widget.List
var localFileItems []string
var currentLocalPath string

var appWindow fyne.Window
var appLogger *logger.Logger

func main() {
	a := app.New()
	appWindow = a.NewWindow("Devcode ObjectStorage Client")

	passwordEntry := widget.NewPasswordEntry()
	dialog.ShowForm("Master Password", "OK", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Password", passwordEntry),
	}, func(ok bool) {
		if !ok {
			appWindow.Close()
			return
		}
		masterPassword = passwordEntry.Text

		mainMenu := createMainMenu(appWindow)
		appWindow.SetMainMenu(mainMenu)
		appWindow.SetContent(createMainUI(appWindow))

		var err error
		currentBookmarks, err = bookmarks.LoadBookmarks(masterPassword)
		if err != nil {
			appLogger.Error(fmt.Errorf("failed to load bookmarks: %w", err))
		} else {
			appLogger.Info("Bookmarks loaded successfully.")
		}

		appWindow.Resize(fyne.NewSize(800, 600))
	}, appWindow)

	appWindow.ShowAndRun()
}

func createMainUI(w fyne.Window) fyne.CanvasObject {
	logLabel := widget.NewLabel("Logs will be displayed here.")
	logLabel.Wrapping = fyne.TextWrapWord
	appLogger = logger.New(logLabel)
	logContainer := container.NewScroll(logLabel)


	// Local file view
	currentLocalPath, _ = os.Getwd()
	localFiles = widget.NewList(
		func() int { return len(localFileItems) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(localFileItems[i]) },
	)
	updateLocalFiles()

	// Remote file view
	remoteFiles = widget.NewList(
		func() int { return len(remoteFileItems) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(remoteFileItems[i]) },
	)

	// Action buttons
	uploadButton := widget.NewButton("Upload", func() { handleUpload(w) })
	downloadButton := widget.NewButton("Download", func() { handleDownload(w) })
	deleteButton := widget.NewButton("Delete", func() { handleDelete(w) })

	actionButtons := container.NewHBox(uploadButton, downloadButton, deleteButton)
	fileView := container.NewHSplit(container.NewBorder(nil, nil, nil, nil, localFiles), remoteFiles)

	mainView := container.NewBorder(nil, actionButtons, nil, nil, container.NewVSplit(fileView, logContainer))

	return mainView
}

func updateLocalFiles() {
	localFileItems = []string{}
	files, err := os.ReadDir(currentLocalPath)
	if err != nil {
		if appLogger != nil {
			appLogger.Error(fmt.Errorf("error reading local directory: %w", err))
		}
		return
	}
	for _, f := range files {
		localFileItems = append(localFileItems, f.Name())
	}
	localFiles.Refresh()
	if appLogger != nil {
		appLogger.Info("Local file list updated for path: " + currentLocalPath)
	}
}

func handleUpload(w fyne.Window) {
	if s3Client == nil || activeBookmark == nil || activeBookmark.BucketName == "" {
		dialog.ShowInformation("Error", "Not connected to a bucket.", w)
		return
	}
	if len(localFiles.Selected) == 0 {
		dialog.ShowInformation("Error", "No local file selected.", w)
		return
	}

	selectedFileName := localFileItems[localFiles.Selected[0]]
	localPath := filepath.Join(currentLocalPath, selectedFileName)

	appLogger.Info(fmt.Sprintf("Uploading '%s' to '%s'...", selectedFileName, activeBookmark.BucketName))
	err := s3Client.UploadFile(activeBookmark.BucketName, selectedFileName, localPath)
	if err != nil {
		appLogger.Error(err)
		dialog.ShowError(err, w)
		return
	}
	appLogger.Info("Upload successful.")
	updateRemoteViewWithBucketObjects(w, activeBookmark.BucketName, activeBookmark.DefaultPrefix)
}

func handleDownload(w fyne.Window) {
	if s3Client == nil || activeBookmark == nil || activeBookmark.BucketName == "" {
		dialog.ShowInformation("Error", "Not connected to a bucket.", w)
		return
	}
	if len(remoteFiles.Selected) == 0 {
		dialog.ShowInformation("Error", "No remote file selected.", w)
		return
	}

	selectedObjectKey := remoteFileItems[remoteFiles.Selected[0]]
	dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
		if err != nil {
			appLogger.Error(err)
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			return // User cancelled
		}
		defer uri.Close()

		appLogger.Info(fmt.Sprintf("Downloading '%s' to '%s'...", selectedObjectKey, uri.URI().Path()))
		err = s3Client.DownloadFile(activeBookmark.BucketName, selectedObjectKey, uri.URI().Path())
		if err != nil {
			appLogger.Error(err)
			dialog.ShowError(err, w)
		} else {
			appLogger.Info("Download successful.")
		}
	}, w)
}

func handleDelete(w fyne.Window) {
	if s3Client == nil || activeBookmark == nil || activeBookmark.BucketName == "" {
		dialog.ShowInformation("Error", "Not connected to a bucket.", w)
		return
	}
	if len(remoteFiles.Selected) == 0 {
		dialog.ShowInformation("Error", "No remote file selected.", w)
		return
	}

	selectedObjectKey := remoteFileItems[remoteFiles.Selected[0]]
	dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete "+selectedObjectKey+"?", func(ok bool) {
		if ok {
			appLogger.Info(fmt.Sprintf("Deleting '%s'...", selectedObjectKey))
			err := s3Client.DeleteObject(activeBookmark.BucketName, selectedObjectKey)
			if err != nil {
				appLogger.Error(err)
				dialog.ShowError(err, w)
				return
			}
			appLogger.Info("Delete successful.")
			updateRemoteViewWithBucketObjects(w, activeBookmark.BucketName, activeBookmark.DefaultPrefix)
		}
	}, w)
}


func createMainMenu(w fyne.Window) *fyne.MainMenu {
	manageBookmarksItem := fyne.NewMenuItem("Manage Bookmarks...", func() {
		showBookmarkDialog(w)
	})
	bookmarkMenu := fyne.NewMenu("Bookmarks", manageBookmarksItem)
	return fyne.NewMainMenu(bookmarkMenu)
}

func showBookmarkDialog(w fyne.Window) {
	bookmarkList = widget.NewList(
		func() int { return len(currentBookmarks) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(currentBookmarks[i].Name) },
	)

	var d dialog.Dialog
	connectButton := widget.NewButton("Connect", func() {
		if len(bookmarkList.Selected) > 0 {
			selected := currentBookmarks[bookmarkList.Selected[0]]
			activeBookmark = &selected
			connectToBookmark(w, selected)
			d.Hide()
		}
	})

	addButton := widget.NewButton("Add", func() {
		showAddEditBookmarkDialog(w, nil, func(newBookmark bookmarks.Bookmark) {
			currentBookmarks = append(currentBookmarks, newBookmark)
			saveBookmarks()
			bookmarkList.Refresh()
			appLogger.Info(fmt.Sprintf("Added new bookmark: %s", newBookmark.Name))
		})
	})

	editButton := widget.NewButton("Edit", func() {
		if len(bookmarkList.Selected) > 0 {
			selectedIndex := bookmarkList.Selected[0]
			showAddEditBookmarkDialog(w, &currentBookmarks[selectedIndex], func(updatedBookmark bookmarks.Bookmark) {
				currentBookmarks[selectedIndex] = updatedBookmark
				saveBookmarks()
				bookmarkList.Refresh()
				appLogger.Info(fmt.Sprintf("Updated bookmark: %s", updatedBookmark.Name))
			})
		}
	})

	deleteButton := widget.NewButton("Delete", func() {
		if len(bookmarkList.Selected) > 0 {
			selectedIndex := bookmarkList.Selected[0]
			bookmarkName := currentBookmarks[selectedIndex].Name
			currentBookmarks = append(currentBookmarks[:selectedIndex], currentBookmarks[selectedIndex+1:]...)
			saveBookmarks()
			bookmarkList.Refresh()
			appLogger.Info(fmt.Sprintf("Deleted bookmark: %s", bookmarkName))
		}
	})

	buttons := container.NewHBox(connectButton, addButton, editButton, deleteButton)
	content := container.NewBorder(nil, buttons, nil, nil, bookmarkList)

	d = dialog.NewCustom("Manage Bookmarks", "Close", content, w)
	d.Show()
}

func connectToBookmark(w fyne.Window, b bookmarks.Bookmark) {
	appLogger.Info(fmt.Sprintf("Connecting to '%s'...", b.Name))
	var err error
	s3Client, err = s3client.New(b)
	if err != nil {
		appLogger.Error(err)
		dialog.ShowError(err, w)
		return
	}
	appLogger.Info("Connection successful.")

	if b.BucketName != "" {
		updateRemoteViewWithBucketObjects(w, b.BucketName, b.DefaultPrefix)
	} else {
		updateRemoteViewWithBuckets(w)
	}
}

func updateRemoteViewWithBuckets(w fyne.Window) {
	appLogger.Info("Listing buckets...")
	buckets, err := s3Client.ListBuckets()
	if err != nil {
		appLogger.Error(err)
		dialog.ShowError(err, w)
		return
	}
	remoteFileItems = buckets
	remoteFiles.Refresh()
	appLogger.Info(fmt.Sprintf("Found %d buckets.", len(buckets)))
}

func updateRemoteViewWithBucketObjects(w fyne.Window, bucket, prefix string) {
	appLogger.Info(fmt.Sprintf("Listing objects in bucket '%s' with prefix '%s'...", bucket, prefix))
	objects, err := s3Client.ListObjects(bucket, prefix)
	if err != nil {
		appLogger.Error(err)
		dialog.ShowError(err, w)
		return
	}
	remoteFileItems = objects
	remoteFiles.Refresh()
	appLogger.Info(fmt.Sprintf("Found %d objects.", len(objects)))
}

func showAddEditBookmarkDialog(w fyne.Window, b *bookmarks.Bookmark, onSave func(bookmarks.Bookmark)) {
	nameEntry := widget.NewEntry()
	endpointEntry := widget.NewEntry()
	bucketEntry := widget.NewEntry()
	keyEntry := widget.NewEntry()
	secretEntry := widget.NewPasswordEntry()
	regionEntry := widget.NewEntry()
	prefixEntry := widget.NewEntry()

	if b != nil {
		nameEntry.SetText(b.Name)
		endpointEntry.SetText(b.Endpoint)
		bucketEntry.SetText(b.BucketName)
		keyEntry.SetText(b.ClientKey)
		secretEntry.SetText(b.SecretKey)
		regionEntry.SetText(b.Region)
		prefixEntry.SetText(b.DefaultPrefix)
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Endpoint", endpointEntry),
		widget.NewFormItem("Bucket Name", bucketEntry),
		widget.NewFormItem("Client Key", keyEntry),
		widget.NewFormItem("Secret Key", secretEntry),
		widget.NewFormItem("Region", regionEntry),
		widget.NewFormItem("Default Prefix", prefixEntry),
	}

	dialog.ShowForm("Bookmark", "Save", "Cancel", formItems, func(ok bool) {
		if ok {
			newBookmark := bookmarks.Bookmark{
				Name:          nameEntry.Text,
				Endpoint:      endpointEntry.Text,
				BucketName:    bucketEntry.Text,
				ClientKey:     keyEntry.Text,
				SecretKey:     secretEntry.Text,
				Region:        regionEntry.Text,
				DefaultPrefix: prefixEntry.Text,
			}
			onSave(newBookmark)
		}
	}, w)
}

func saveBookmarks() {
	err := bookmarks.SaveBookmarks(currentBookmarks, masterPassword)
	if err != nil {
		appLogger.Error(fmt.Errorf("failed to save bookmarks: %w", err))
	} else {
		appLogger.Info("Bookmarks saved successfully.")
	}
}