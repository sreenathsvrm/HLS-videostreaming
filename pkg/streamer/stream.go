package streamer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Stream(c *gin.Context) {
	// Fetch video id and playlist name from path parameters
	videoID := c.Param("video_id")
	playlist := c.Param("playlist")

	// Create a channel to receive the playlist data
	playlistDataChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		// Fetch the playlist data in a separate goroutine
		playlistData, err := readPlaylistData(videoID, playlist)
		if err != nil {
			errChan <- err // Send the error through the error channel
			return
		}
		playlistDataChan <- playlistData // Send the playlist data through the data channel
	}()

	select { //select helps to perform non-blocking operation on channels. It waits until one case is ready to proceed, and if multiple cases are ready to proceed, it randomly selects one to execure.
	case playlistData := <-playlistDataChan: // if data is available on the playlistDataChan channel
		// Set the response headers
		c.Header("Content-Type", "application/vnd.apple.mpegurl")
		c.Header("Content-Disposition", "inline")

		// Write the playlist data to the response body
		c.Writer.Write(playlistData)
	case err := <-errChan: // This case checks if there is an error available on the errChan channel
		// Handle the error
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to read file from server",
			"error":   err.Error(),
		})
	case <-time.After(5 * time.Second): //This case uses the time.After function to create a channel that sends a value after a specified duration.
		// Handle the timeout case
		c.JSON(http.StatusGatewayTimeout, gin.H{
			"message": "request timed out",
		})
	}

}

func readPlaylistData(videoID, playlist string) ([]byte, error) {
	// Construct the playlist file path
	playlistPath := fmt.Sprintf("storage/%s/%s", videoID, playlist)

	// Read the playlist file
	playlistData, err := ioutil.ReadFile(playlistPath)
	if err != nil {
		return nil, err
	}
	return playlistData, nil
}
