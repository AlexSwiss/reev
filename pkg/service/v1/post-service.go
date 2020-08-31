package v1

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/AlexSwiss/reev/pkg/api/v1"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

// PostServiceServer is implementation of v1.PosrServiceServer proto interface
type PostServiceServer struct {
	db *sql.DB
}

// NewPostServiceServer creates Post service
func NewPostServiceServer(db *sql.DB) v1.PostServiceServer {
	return &PostServiceServer{db: db}
}

// checkAPI checks if the API version requested by client is supported by server
func (s *PostServiceServer) checkAPI(api string) error {
	// API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *PostServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
	}
	return c, nil
}

// Create new Post task
func (s *PostServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// insert Post entity data
	res, err := c.ExecContext(ctx, "INSERT INTO Post(`Title`, `Description`) VALUES(?, ?)",
		req.Post.Title, req.Post.Description)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into Post-> "+err.Error())
	}

	// get ID of created Post
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created Post-> "+err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

// read post
func (s *PostServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// query Post by ID
	rows, err := c.QueryContext(ctx, "SELECT `ID`, `Title`, `Description`, `Reminder` FROM Post WHERE `ID`=?",
		req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Post-> "+err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from Post-> "+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Post with ID='%d' is not found",
			req.Id))
	}

	// get Post data
	var td v1.Post
	if err := rows.Scan(&td.Id, &td.Title, &td.Description); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve field values from Post row-> "+err.Error())
	}

	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("found multiple Post rows with ID='%d'",
			req.Id))
	}

	return &v1.ReadResponse{
		Api:  apiVersion,
		Post: &td,
	}, nil

}

// Update post task
func (s *PostServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// update Post
	res, err := c.ExecContext(ctx, "UPDATE Post SET `Title`=?, `Description`=? WHERE `ID`=?",
		req.Post.Title, req.Post.Description, req.Post.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update Post-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Post with ID='%d' is not found",
			req.Post.Id))
	}

	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

// Delete post
func (s *PostServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// delete Post
	res, err := c.ExecContext(ctx, "DELETE FROM Post WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete Post-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Post with ID='%d' is not found",
			req.Id))
	}

	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

// ReadAll post
func (s *PostServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// get Post list
	rows, err := c.QueryContext(ctx, "SELECT `ID`, `Title`, `Description` FROM Post")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Post-> "+err.Error())
	}
	defer rows.Close()

	list := []*v1.Post{}
	for rows.Next() {
		td := new(v1.Post)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Post row-> "+err.Error())
		}
		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Post-> "+err.Error())
	}

	return &v1.ReadAllResponse{
		Api:   apiVersion,
		Posts: list,
	}, nil
}
