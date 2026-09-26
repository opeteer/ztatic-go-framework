package repositories

import (
	"context"

	"github.com/Masterminds/squirrel"
	"ztatic-go-framework/data"
	"myapp/internal/models"
)

type ArticleRepository struct {
	*data.BaseRepository[models.Article]
}

func NewArticleRepository(db *data.DBEngine) *ArticleRepository {
	return &ArticleRepository{
		BaseRepository: data.NewBaseRepository[models.Article](db, "articles"),
	}
}

// Custom query using Squirrel AST query building
func (r *ArticleRepository) FindByAuthor(ctx context.Context, author string) ([]models.Article, error) {
	query, args, err := r.DB.Builder.
		Select("id", "title", "content", "author", "created_at").
		From(r.TableName).
		Where(squirrel.Eq{"author": author}).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.DB.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Author, &a.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
