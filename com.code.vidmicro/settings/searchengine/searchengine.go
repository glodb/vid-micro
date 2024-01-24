package searchengine

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/meilisearch/meilisearch-go"
)

var (
	instance *searchEngine
	once     sync.Once
)

type searchEngine struct {
	client *meilisearch.Client
	index  *meilisearch.Index
}

func GetInstance() *searchEngine {
	once.Do(func() {

		instance = &searchEngine{}

		instance.client = meilisearch.NewClient(meilisearch.ClientConfig{
			Host:   configmanager.GetInstance().Meilisearch.Host,
			APIKey: configmanager.GetInstance().Meilisearch.ApiKey,
		})
		instance.index = instance.createIndex(configmanager.GetInstance().MeilisearchIndex)
	})
	return instance
}

// Create an index if it doesn't exist
func (se *searchEngine) createIndex(indexName string) *meilisearch.Index {
	index := se.client.Index(indexName)
	return index
}

func (se *searchEngine) AddDocuments() {

	// indexName := "titles" // Your index name
	// index, err := client.Indexes().Create(meilisearch.Index{
	// 	Name: indexName,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Add documents to the index
	// titles := []Title{
	// 	{ID: 1, OriginalTitle: "Suits", Year: 2011, CoverURL: "example.com/suits", Languages: []string{"English", "Spanish"}},
	// 	// Add more titles as needed
	// }

	// for _, title := range titles {
	// 	_, err := index.AddDocuments([]Title{title})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// return index

}
