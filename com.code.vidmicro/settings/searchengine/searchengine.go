package searchengine

import (
	"fmt"
	"log"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/bytedance/sonic"
	"github.com/meilisearch/meilisearch-go"
)

var (
	instance *searchEngine
	once     sync.Once
)

type searchEngine struct {
	client *meilisearch.Client
}

func GetInstance() *searchEngine {
	once.Do(func() {

		instance = &searchEngine{}

		instance.client = meilisearch.NewClient(meilisearch.ClientConfig{
			Host:   configmanager.GetInstance().Meilisearch.Host,
			APIKey: configmanager.GetInstance().Meilisearch.ApiKey,
		})
		instance.createIndex(configmanager.GetInstance().MeilisearchIndex)
	})
	return instance
}

// Create an index if it doesn't exist
func (se *searchEngine) createIndex(indexName string) *meilisearch.Index {
	index := se.client.Index(indexName)
	resp, err := index.GetFilterableAttributes()

	if err != nil || len(*resp) <= 0 {
		index.UpdateFilterableAttributes(&[]string{"id", "original_title", "year", "alternnative_name", "type", "genres"})
	}
	return index
}

func (se *searchEngine) ProcessTitleDocuments(title models.MeilisearchTitle) {

	tempTitle := models.MeilisearchTitle{}
	err := se.client.Index(configmanager.GetInstance().MeilisearchIndex).GetDocument(fmt.Sprintf("%d", title.Id), nil, &tempTitle)
	if err != nil {
		log.Println(err)
	}

	tempTitle.Id = title.Id

	if title.OriginalTitle != tempTitle.OriginalTitle && title.OriginalTitle != "" {
		tempTitle.OriginalTitle = title.OriginalTitle
	}

	if title.AlternativeName != tempTitle.AlternativeName && title.AlternativeName != "" {
		tempTitle.AlternativeName = title.AlternativeName
	}

	if title.Year != tempTitle.Year && title.Year > 0 {
		tempTitle.Year = title.Year
	}

	if title.CoverUrl != tempTitle.CoverUrl && title.CoverUrl != "" {
		tempTitle.CoverUrl = title.CoverUrl
	}

	if title.Sequence != tempTitle.Sequence && title.Sequence > 0 {
		tempTitle.Sequence = title.Sequence
	}

	if title.TypeId != tempTitle.TypeId && title.TypeId > 0 {
		tempTitle.TypeId = title.TypeId
	}

	if title.TypeName != tempTitle.TypeName && title.TypeName != "" {
		tempTitle.TypeName = title.TypeName
	}

	if title.Score != tempTitle.Score && title.Score > 0 {
		tempTitle.Score = title.Score
	}

	info, err := se.client.Index(configmanager.GetInstance().MeilisearchIndex).UpdateDocuments([]models.MeilisearchTitle{title})
	if err != nil {
		log.Println(err)
	} else {
		log.Println(info)
	}

}

func (se *searchEngine) SearchDocuments(title models.MeilisearchTitle, pageSize int64, page int64, filter string) (models.PaginationResults, error) {
	offset := pageSize * (page - 1)
	searchRes, err := se.client.Index(configmanager.GetInstance().MeilisearchIndex).Search(title.OriginalTitle, &meilisearch.SearchRequest{
		Filter: filter,
		Limit:  pageSize,
		Offset: offset,
	})

	if err != nil {
		return models.PaginationResults{}, err
	}
	var titles []models.MeilisearchTitle
	for _, hit := range searchRes.Hits {
		var title models.MeilisearchTitle
		// Convert the hit to JSON bytes
		hitJSON, err := sonic.Marshal(hit)
		if err != nil {
			log.Fatal(err)
		}
		// Unmarshal JSON bytes to custom struct
		err = sonic.Unmarshal(hitJSON, &title)
		if err != nil {
			log.Fatal(err)
		}
		titles = append(titles, title)
	}

	pr := models.PaginationResults{
		Pagination: models.NewPagination(searchRes.EstimatedTotalHits, int(pageSize), int(page)),
		Data:       titles,
	}
	return pr, nil
}

func (se *searchEngine) DeleteDocuments(title models.MeilisearchTitle) error {
	title.Id = 287947
	_, err := se.client.Index(configmanager.GetInstance().MeilisearchIndex).DeleteDocument(fmt.Sprintf("%d", title.Id))
	return err
}
