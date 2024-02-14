package content

import (
	"context"
	"geekswimmers/storage"
)

func FindHighlightedArticles(db storage.Database) ([]*Article, error) {
	stmt := `select a.reference, a.title, a.abstract, a.highlighted, a.published, a.content, a.image, a.image_copyright
			 from article a
			 where a.highlighted = true
			 order by a.published desc`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []*Article
	for rows.Next() {
		article := &Article{}
		err = rows.Scan(&article.Reference, &article.Title, &article.Abstract,
			&article.Highlighted, &article.Published, &article.Content, &article.Image, &article.ImageCopyright)

		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func findArticle(reference string, db storage.Database) (*Article, error) {
	stmt := `select a.reference, a.title, a.abstract, a.published, a.content, a.image, a.image_copyright
			 from article a
			 where a.reference = $1`

	row := db.QueryRow(context.Background(), stmt, reference)

	article := &Article{}
	err := row.Scan(&article.Reference, &article.Title, &article.Abstract, &article.Published, &article.Content, &article.Image, &article.ImageCopyright)
	if err != nil {
		return nil, err
	}

	return article, nil
}
