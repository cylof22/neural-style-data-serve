package main

// ImageStore define the basic interface for a ImageStore
type ImageStore interface {
	Save(image Image) (string, error)
	Find(userID, imgName string) (string, error)
	FindAllByUser(userID string) ([]string, error)
}
