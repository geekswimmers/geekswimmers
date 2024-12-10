package content

import (
	"context"
	"fmt"
	"geekswimmers/storage"
	"os"
)

func FindHighlightedArticles(db storage.Database) ([]*Article, error) {
	sql := `select a.reference, a.title, a.abstract, a.highlighted, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
			from article a
			where a.highlighted = true
			order by a.published desc`
	rows, err := db.Query(context.Background(), sql)
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

func FindArticlesExcept(reference string, db storage.Database) ([]*Article, error) {
	sql := `select a.reference, a.title, coalesce(a.sub_title, ''), a.abstract, a.highlighted, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
			from article a
			where a.reference != $1
			order by a.published desc`
	rows, err := db.Query(context.Background(), sql, reference)
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

func getArticle(reference string, db storage.Database) (*Article, error) {
	sql := `select a.reference, a.title, a.abstract, a.published, a.content, coalesce(a.image, ''), coalesce(a.image_copyright, '')
			 from article a
			 where a.reference = $1`

	row := db.QueryRow(context.Background(), sql, reference)

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

func GetQuoteOfTheDay(dayOfYear int, db storage.Database) (*Quote, error) {
	sql := `select count(seq) from quote`
	row := db.QueryRow(context.Background(), sql)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("GetQuoteOfTheDay.count: %v", err)
	}

	if count > 0 {
		sql = `select seq, quote, coalesce(author, '')
				from quote q
				where q.seq = $1`

		seq := getQuoteSequence(dayOfYear, count)
		row = db.QueryRow(context.Background(), sql, seq)

		quote := &Quote{}
		err = row.Scan(&quote.Sequence, &quote.Quote, &quote.Author)
		if err != nil {
			return nil, fmt.Errorf("GetQuoteOfTheDay.quotes: seq: %v, count: %v, dayOfYear: %v, error: %v", seq, count, dayOfYear, err)
		}
		return quote, nil
	}

	return nil, nil
}

func getQuoteSequence(dayOfYear, count int) int {
	seq := (dayOfYear - 1) % count

	if seq == 0 {
		seq = count
	}

	return seq
}

func FindUpdates(db storage.Database) ([]*ServiceUpdate, error) {
	sql := `select su.title, su.content, su.published
			from service_update su
			order by su.published desc
			limit 5`
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []*ServiceUpdate
	for rows.Next() {
		update := &ServiceUpdate{}
		err = rows.Scan(&update.Title, &update.Content, &update.Published)

		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
		}
		updates = append(updates, update)
	}

	return updates, nil
}
