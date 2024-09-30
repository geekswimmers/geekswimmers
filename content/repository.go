package content

import (
	"context"
	"fmt"
	"geekswimmers/storage"
	"os"
)

func FindHighlightedArticles(db storage.Database) ([]*Article, error) {
	stmt := `select a.reference, a.title, a.abstract, a.highlighted, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
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

		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func findArticlesExcept(reference string, db storage.Database) ([]*Article, error) {
	stmt := `select a.reference, a.title, coalesce(a.sub_title, ''), a.abstract, a.highlighted, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
			 from article a
			 where a.reference != $1
			 order by a.published desc`
	rows, err := db.Query(context.Background(), stmt, reference)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []*Article
	for rows.Next() {
		article := &Article{}
		err = rows.Scan(&article.Reference, &article.Title, &article.SubTitle, &article.Abstract,
			&article.Highlighted, &article.Published, &article.Content, &article.Image, &article.ImageCopyright)

		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func findArticle(reference string, db storage.Database) (*Article, error) {
	stmt := `select a.reference, a.title, a.abstract, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
			 from article a
			 where a.reference = $1`

	row := db.QueryRow(context.Background(), stmt, reference)

	article := &Article{}
	err := row.Scan(&article.Reference, &article.Title, &article.Abstract, &article.Published, &article.Content, &article.Image, &article.ImageCopyright)
	if err != nil {
		return nil, err
	}

	article.Content, err = loadContent(fmt.Sprintf("web/content/%s", article.Content))
	if err != nil {
		return nil, err
	}

	return article, nil
}

func loadContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetQuoteOfTheDay(dayOfYear int64, db storage.Database) (*Quote, error) {
	stmt := `select count(seq) from quote`
	row := db.QueryRow(context.Background(), stmt)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		stmt = `select seq, quote, coalesce(author, '')
				from quote q
				where q.seq = $1`

		seq := (dayOfYear - 1) % count

		row = db.QueryRow(context.Background(), stmt, seq)

		quote := &Quote{}
		err = row.Scan(&quote.Sequence, &quote.Quote, &quote.Author)
		if err != nil {
			return nil, err
		}
		return quote, nil
	}

	return nil, nil
}
