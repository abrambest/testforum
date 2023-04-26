package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/saparaly/snippentbox/pkg/models"
)

type UserModel struct {
	DB *sql.DB
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
func (m *UserModel) Insert(user_id int, title, content, username, tags string) (int, error) {
	stmt := `INSERT INTO posts (user_id, title, tags, username, content, created)
	VALUES(?, ?, ?, ?, ?, datetime('now','utc'))`

	result, err := m.DB.Exec(stmt, user_id, title, tags, username, content)
	if err != nil {
		fmt.Println("111111111111111111111111111111111111111111111")
		return 0, nil
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *UserModel) Get(id int) (*models.Post, error) {
	stmt := `SELECT id, title, username, content, tags, like, dislike, created FROM posts WHERE id=?`

	row := m.DB.QueryRow(stmt, id)

	s := &models.Post{}

	err := row.Scan(&s.Id, &s.Title, &s.UserName, &s.Description, &s.Tags, &s.Like, &s.Dislike, &s.Created)
	// fmt.Println(err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

// This will return the 10 most recently created snippets.
func (m *UserModel) Latest() ([]*models.Post, error) {
	stmt := `SELECT id, user_id, title, content, tags, created FROM posts`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		s := &models.Post{}

		err = rows.Scan(&s.Id, &s.AuthorId, &s.Title, &s.Description, &s.Tags, &s.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (m *UserModel) InsertComment(comment *models.Comment) error {
	stmt := "INSERT INTO comments (content, user_id, username, post_id) VALUES (?, ?, ?, ?)"

	result, err := m.DB.Exec(stmt, comment.Text, comment.UserId, comment.UserName, comment.PostId)
	if err != nil {
		return err
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	comment.Id = int(lastInsertId)

	return nil
}

func (m *UserModel) GetComments(postID int) ([]*models.Comment, error) {
	// Prepare the SQL query to retrieve all comments for a post.
	query := `
        SELECT id, post_id, user_id, content, username, like, dislike
        FROM comments
        WHERE post_id = $1
    `

	// Execute the query and retrieve the results.
	rows, err := m.DB.Query(query, postID)
	if err != nil {
		fmt.Println("error4")
		return nil, err
	}
	defer rows.Close()

	// Loop through the rows and build the slice of comments.
	comments := []*models.Comment{}
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(&comment.Id, &comment.PostId, &comment.UserId, &comment.Text, &comment.UserName, &comment.Like, &comment.Dislike)
		if err != nil {
			fmt.Println("error5")
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("error5")
		return nil, err
	}

	return comments, nil
}

func (m *UserModel) LikePost(postID int, userID int) error {
	// Check if user has already liked the post
	row := m.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = $1 AND user_id = $2", postID, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// User has already liked the post, remove the like
		_, err := m.DB.Exec("DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2", postID, userID)
		if err != nil {
			return err
		}
	} else {
		// User has not yet liked the post, add the like
		_, err := m.DB.Exec("INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)", postID, userID)
		if err != nil {
			return err
		}
	}

	// Get the like count for the post
	row = m.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = $1", postID)
	var likeCount int
	err = row.Scan(&likeCount)
	if err != nil {
		return err
	}

	// Remove the dislike if there is any
	_, err = m.DB.Exec("DELETE FROM post_dislikes WHERE post_id = $1 AND user_id = $2", postID, userID)
	if err != nil {
		return err
	}

	// Update the post record with the new like and dislike counts
	_, err = m.DB.Exec("UPDATE posts SET like = $1, dislike = (SELECT COUNT(*) FROM post_dislikes WHERE post_id = $2) WHERE id = $2", likeCount, postID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) DislikePost(postID int, userID int) error {
	// Check if user has already liked the post
	row := m.DB.QueryRow("SELECT COUNT(*) FROM post_dislikes WHERE post_id = $1 AND user_id = $2", postID, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// User has already liked the post, remove the like
		_, err := m.DB.Exec("DELETE FROM post_dislikes WHERE post_id = $1 AND user_id = $2", postID, userID)
		if err != nil {
			return err
		}
	} else {
		// User has not yet liked the post, add the like
		_, err := m.DB.Exec("INSERT INTO post_dislikes (post_id, user_id) VALUES ($1, $2)", postID, userID)
		if err != nil {
			return err
		}

	}

	// Get the like count for the post
	row = m.DB.QueryRow("SELECT COUNT(*) FROM post_dislikes WHERE post_id = $1", postID)
	var likeCount int
	err = row.Scan(&likeCount)
	if err != nil {
		return err
	}

	// Remove the like if there is any
	_, err = m.DB.Exec("DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2", postID, userID)
	if err != nil {
		return err
	}
	// Update the post record with the new like count
	_, err = m.DB.Exec("UPDATE posts SET dislike = $1, like = (SELECT COUNT(*) FROM post_likes WHERE post_id = $2) WHERE id = $2", likeCount, postID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) LikeComment(commentID int, userID int) error {
	// Check if user has already liked the comment
	row := m.DB.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// User has already liked the comment, remove the like
		_, err := m.DB.Exec("DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
		if err != nil {
			return err
		}
	} else {
		// User has not yet liked the comment, add the like
		_, err := m.DB.Exec("INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)", commentID, userID)
		if err != nil {
			return err
		}
	}

	// Get the like count for the comment
	row = m.DB.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1", commentID)
	var likeCount int
	err = row.Scan(&likeCount)
	if err != nil {
		return err
	}

	// Remove the dislike if there is any
	_, err = m.DB.Exec("DELETE FROM comment_dislikes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		return err
	}

	// Update the comment record with the new like and dislike counts
	_, err = m.DB.Exec("UPDATE comments SET like = $1, dislike = (SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = $2) WHERE id = $2", likeCount, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) DislikeComment(commentID int, userID int) error {
	// Check if user has already liked the comment
	row := m.DB.QueryRow("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// User has already liked the comment, remove the like
		_, err := m.DB.Exec("DELETE FROM comment_dislikes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
		if err != nil {
			return err
		}
	} else {
		// User has not yet liked the comment, add the like
		_, err := m.DB.Exec("INSERT INTO comment_dislikes (comment_id, user_id) VALUES ($1, $2)", commentID, userID)
		if err != nil {
			return err
		}
	}

	// Get the like count for the comment
	row = m.DB.QueryRow("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = $1", commentID)
	var likeCount int
	err = row.Scan(&likeCount)
	if err != nil {
		return err
	}

	// Remove the like if there is any
	_, err = m.DB.Exec("DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		return err
	}
	// Update the comment record with the new like and dislike counts
	_, err = m.DB.Exec("UPDATE comments SET dislike = $1, like = (SELECT COUNT(*) FROM comment_likes WHERE comment_id = $2) WHERE id = $2", likeCount, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) GetLikedPosts(userID int) ([]*models.Post, error) {
	// Get the IDs of all posts that have been liked by the user
	rows, err := m.DB.Query("SELECT post_id FROM post_likes WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a slice to hold the liked posts
	likedPosts := []*models.Post{}

	// Iterate through the rows and add each post to the likedPosts slice
	for rows.Next() {
		var postID int
		err := rows.Scan(&postID)
		if err != nil {
			return nil, err
		}

		// Use the Get method to get the post by its ID
		post, err := m.Get(postID)
		if err != nil {
			return nil, err
		}

		likedPosts = append(likedPosts, post)
	}

	// Check for any errors that occurred while iterating through the rows
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return likedPosts, nil
}

func (m *UserModel) GetPostsByTag(tag string) ([]*models.Post, error) {
	stmt := `SELECT id, title, content, tags, created
             FROM posts WHERE tags LIKE '%' || ? || '%'`

	rows, err := m.DB.Query(stmt, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}
	for rows.Next() {
		p := &models.Post{}
		err := rows.Scan(&p.Id, &p.Title, &p.Description, &p.Tags, &p.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *UserModel) GetUserIDByToken(token string) (*models.User, error) {
	/**/
	stmt := `SELECT u.id, u.email, u.name, u.password FROM session s JOIN users u ON s.UserID = u.id WHERE s.Token = ?`

	row := m.DB.QueryRow(stmt, token)
	user := &models.User{}

	err := row.Scan(&user.Id, &user.Email, &user.Username, &user.Password)
	// fmt.Println(err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (m *UserModel) GetUserCreatedPost(userID int) ([]*models.Post, error) {
	stmt := `SELECT id, title, username, content, tags, created
	         FROM posts WHERE user_id = ?`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}

	for rows.Next() {
		p := &models.Post{}
		err := rows.Scan(&p.Id, &p.Title, &p.UserName, &p.Description, &p.Tags, &p.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
