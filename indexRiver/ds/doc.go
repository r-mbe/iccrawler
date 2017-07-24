
package ds

    // The first goroutine will emit documents and send it to the second goroutine
    // via the docsc channel.
    // The second Goroutine will simply bulk insert the documents.
    type DelDoc struct {
        Index     string    `json:"index"`
        Type      string    `json:"type"`
        ID        string    `json:"id"`
       // Timestamp time.Time `json:"@timestamp"`
    }


