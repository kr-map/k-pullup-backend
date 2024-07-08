package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Alfex4936/chulbong-kr/dto"
	"github.com/Alfex4936/chulbong-kr/model"
	"github.com/iancoleman/orderedmap"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB        *sqlx.DB
	S3Service *S3Service
}

func NewUserService(db *sqlx.DB, s3Service *S3Service) *UserService {
	return &UserService{
		DB:        db,
		S3Service: s3Service,
	}
}

// GetUserById retrieves a user by their email address
func (s *UserService) GetUserById(userID int) (*dto.UserResponse, error) {
	var user dto.UserResponse

	// Define the query to select the user
	query := `SELECT UserID, Username, Email FROM Users WHERE UserID = ?`

	// Execute the query
	err := s.DB.Get(&user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user found with userID %d", userID)
		}
		return nil, fmt.Errorf("error fetching user by userID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User

	// Define the query to select the user
	query := `SELECT UserID, Username, Email, PasswordHash, Provider, ProviderID, CreatedAt, UpdatedAt FROM Users WHERE Email = ?`

	// Execute the query
	err := s.DB.Get(&user, query, email)
	if err != nil {
		return nil, err
		// if err == sql.ErrNoRows {
		// 	// No user found with the provided email
		// 	return nil, fmt.Errorf("no user found with email %s", email)
		// }
		// // An error occurred during the query execution
		// return nil, fmt.Errorf("error fetching user by email: %w", err)
	}

	return &user, nil
}

func (s *UserService) UpdateUserProfile(userID int, updateReq *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if updateReq.Username != nil {
		normalizedUsername := strings.TrimSpace(SegmentConsonants(*updateReq.Username))
		var existingID int
		err = tx.Get(&existingID, "SELECT UserID FROM Users WHERE Username = ?", normalizedUsername)
		if err == nil || err != sql.ErrNoRows {
			return nil, fmt.Errorf("username %s is already in use", normalizedUsername)
		}
		*updateReq.Username = normalizedUsername
	}

	if updateReq.Email != nil {
		var existingID int
		err = tx.Get(&existingID, "SELECT UserID FROM Users WHERE Email = ?", *updateReq.Email)
		if err == nil || err != sql.ErrNoRows {
			return nil, fmt.Errorf("email %s is already in use", *updateReq.Email)
		}
	}

	var setParts []string
	var args []any

	if updateReq.Username != nil {
		setParts = append(setParts, "Username = ?")
		args = append(args, *updateReq.Username)
	}

	if updateReq.Email != nil {
		setParts = append(setParts, "Email = ?")
		args = append(args, *updateReq.Email)
	}

	if updateReq.Password != nil {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(*updateReq.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, hashErr
		}
		setParts = append(setParts, "PasswordHash = ?")
		args = append(args, string(hashedPassword))
	}

	if len(setParts) > 0 {
		args = append(args, userID)
		query := fmt.Sprintf("UPDATE Users SET %s WHERE UserID = ?", strings.Join(setParts, ", "))
		_, err = tx.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("error updating user: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing update: %w", err)
	}

	// Fetch the updated user details
	updatedUser, err := s.GetUserById(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching updated user: %w", err)
	}

	return updatedUser, nil
}

// GetAllReportsByUser retrieves all reports submitted by a specific user from the database.
func (s *UserService) GetAllReportsByUser(userID int) ([]dto.MarkerReportResponse, error) {
	const query = `
    SELECT r.ReportID, r.MarkerID, r.UserID, ST_X(r.Location) AS Latitude, ST_Y(r.Location) AS Longitude,
    ST_X(r.NewLocation) AS NewLatitude, ST_Y(r.NewLocation) AS NewLongitude,
    r.Description, r.CreatedAt, r.Status, r.DoesExist, m.Address, p.PhotoURL
    FROM Reports r
    LEFT JOIN ReportPhotos p ON r.ReportID = p.ReportID
	LEFT JOIN Markers m ON r.MarkerID = m.MarkerID
    WHERE r.UserID = ?
    ORDER BY r.CreatedAt DESC
    `
	rows, err := s.DB.Queryx(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying reports by user: %w", err)
	}
	defer rows.Close()

	reportMap := make(map[int]*dto.MarkerReportResponse)
	for rows.Next() {
		var (
			r   dto.MarkerReportResponse
			url sql.NullString // Use sql.NullString to handle possible NULL values from PhotoURL
		)
		if err := rows.Scan(&r.ReportID, &r.MarkerID, &r.UserID, &r.Latitude, &r.Longitude,
			&r.NewLatitude, &r.NewLongitude, &r.Description, &r.CreatedAt, &r.Status, &r.DoesExist, &r.Address, &url); err != nil {
			return nil, err
		}
		if report, exists := reportMap[r.ReportID]; exists {
			// Append only if url is valid to avoid appending empty strings for reports without photos
			if url.Valid {
				report.PhotoURLs = append(report.PhotoURLs, url.String)
			}
		} else {
			r.PhotoURLs = make([]string, 0)
			if url.Valid {
				r.PhotoURLs = append(r.PhotoURLs, url.String)
			}
			reportMap[r.ReportID] = &r
		}
	}

	// Convert map to slice
	reports := make([]dto.MarkerReportResponse, 0, len(reportMap))
	for _, report := range reportMap {
		reports = append(reports, *report)
	}

	return reports, nil
}

// GetAllReportsForMyMarkersByUser retrieves all reports for markers owned by a specific user
func (s *UserService) GetAllReportsForMyMarkersByUser(userID int) (dto.GroupedReportsResponse, error) {
	const query = `
        SELECT 
            r.ReportID,
            r.MarkerID,
            r.UserID,
            ST_X(r.Location) as Latitude,
            ST_Y(r.Location) as Longitude,
            ST_X(r.NewLocation) as NewLatitude,
            ST_Y(r.NewLocation) as NewLongitude,
            r.Description,
            r.CreatedAt,
            r.Status,
            r.DoesExist,
			m.Address,
            rp.PhotoURL
        FROM 
            Reports r
        LEFT JOIN 
            ReportPhotos rp ON r.ReportID = rp.ReportID
		LEFT JOIN
			Markers m ON r.MarkerID = m.MarkerID
        WHERE 
            EXISTS (
				SELECT 1
				FROM Markers
				WHERE Markers.MarkerID = r.MarkerID
				AND Markers.UserID = ?
			)
        ORDER BY
            r.MarkerID, r.CreatedAt DESC;
    `

	rows, err := s.DB.Queryx(query, userID)
	if err != nil {
		return dto.GroupedReportsResponse{}, fmt.Errorf("error querying reports by user: %w", err)
	}
	defer rows.Close()

	groupedReports := make(map[int][]dto.ReportWithPhotos, 0)
	reportMap := make(map[int]*dto.MarkerReportResponse)
	// Map to track if address is already added for a marker
	addressAdded := make(map[int]struct{})

	for rows.Next() {
		var r dto.MarkerReportResponse
		var url sql.NullString
		if err := rows.Scan(&r.ReportID, &r.MarkerID, &r.UserID, &r.Latitude, &r.Longitude,
			&r.NewLatitude, &r.NewLongitude, &r.Description, &r.CreatedAt, &r.Status, &r.DoesExist, &r.Address, &url); err != nil {
			return dto.GroupedReportsResponse{}, err
		}

		report, exists := reportMap[r.ReportID]
		if exists {
			if url.Valid {
				report.PhotoURLs = append(report.PhotoURLs, url.String)
			}
		} else {
			r.PhotoURLs = make([]string, 0)
			if url.Valid {
				r.PhotoURLs = append(r.PhotoURLs, url.String)
			}
			reportMap[r.ReportID] = &r

			// Add address only if it's the first report for the marker
			reportWithPhotos := dto.ReportWithPhotos{
				ReportID:     r.ReportID,
				Description:  r.Description,
				Status:       r.Status,
				CreatedAt:    r.CreatedAt,
				Photos:       r.PhotoURLs,
				NewLatitude:  r.NewLatitude,
				NewLongitude: r.NewLongitude,
			}
			if _, added := addressAdded[r.MarkerID]; !added {
				reportWithPhotos.Address = r.Address
				addressAdded[r.MarkerID] = struct{}{}
			}
			groupedReports[r.MarkerID] = append(groupedReports[r.MarkerID], reportWithPhotos)
		}
	}

	// Sort each group by status and CreatedAt
	for _, reports := range groupedReports {
		sort.SliceStable(reports, func(i, j int) bool {
			if reports[i].Status == "PENDING" && reports[j].Status != "PENDING" {
				return true
			}
			if reports[i].Status != "PENDING" && reports[j].Status == "PENDING" {
				return false
			}
			return reports[i].CreatedAt.After(reports[j].CreatedAt)
		})
	}

	// Create a slice of marker IDs to sort by the most recent report date
	type MarkerWithLatestReport struct {
		MarkerID   int
		LatestDate time.Time
	}

	markersWithLatestReports := make([]MarkerWithLatestReport, 0, len(groupedReports))
	for markerID, reports := range groupedReports {
		if len(reports) > 0 {
			markersWithLatestReports = append(markersWithLatestReports, MarkerWithLatestReport{
				MarkerID:   markerID,
				LatestDate: reports[0].CreatedAt,
			})
		}
	}

	// Sort markers by the date of their latest report
	sort.SliceStable(markersWithLatestReports, func(i, j int) bool {
		return markersWithLatestReports[i].LatestDate.After(markersWithLatestReports[j].LatestDate)
	})

	// Construct the response in the sorted order of markers
	sortedGroupedReports := orderedmap.New()
	for _, marker := range markersWithLatestReports {
		sortedGroupedReports.Set(strconv.Itoa(marker.MarkerID), groupedReports[marker.MarkerID])
	}

	response := dto.GroupedReportsResponse{
		TotalReports: len(reportMap),
		Markers:      sortedGroupedReports,
	}

	return response, nil
}

func (s *UserService) GetAllFavorites(userID int) ([]dto.MarkerSimpleWithDescrption, error) {
	favorites := make([]dto.MarkerSimpleWithDescrption, 0)
	const query = `
    SELECT Markers.MarkerID, ST_X(Markers.Location) AS Latitude, ST_Y(Markers.Location) AS Longitude, Markers.Description, Markers.Address
    FROM Favorites
    JOIN Markers ON Favorites.MarkerID = Markers.MarkerID
    WHERE Favorites.UserID = ?
    ORDER BY Markers.CreatedAt DESC` // Order by CreatedAt in descending order

	err := s.DB.Select(&favorites, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching favorites: %w", err)
	}

	return favorites, nil
}

// DeleteUserWithRelatedData
func (s *UserService) DeleteUserWithRelatedData(ctx context.Context, userID int) error {
	// Begin a transaction
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	// Fetch Photo URLs associated with the user
	var photoURLs []string
	fetchPhotosQuery := `SELECT PhotoURL FROM Photos WHERE MarkerID IN (SELECT MarkerID FROM Markers WHERE UserID = ?)`
	if err := tx.SelectContext(ctx, &photoURLs, fetchPhotosQuery, userID); err != nil {
		tx.Rollback()
		return fmt.Errorf("fetching photo URLs: %w", err)
	}

	// Delete each photo from S3
	for _, url := range photoURLs {
		if err := s.S3Service.DeleteDataFromS3(url); err != nil {
			tx.Rollback()
			return fmt.Errorf("deleting photo from S3: %w", err)
		}
	}

	// Note: Order matters due to foreign key constraints
	deletionQueries := []string{
		"DELETE FROM OpaqueTokens WHERE UserID = ?",
		"DELETE FROM Comments WHERE UserID = ?",
		"DELETE FROM MarkerDislikes WHERE UserID = ?",
		"DELETE FROM Photos WHERE MarkerID IN (SELECT MarkerID FROM Markers WHERE UserID = ?)",
		"UPDATE Markers SET UserID = NULL WHERE UserID = ?", // Set UserID to NULL for Markers instead of deleting
		"DELETE FROM Users WHERE UserID = ?",
	}

	// Execute each deletion query within the transaction
	for _, query := range deletionQueries {
		if _, err := tx.ExecContext(ctx, query, userID); err != nil {
			tx.Rollback() // Attempt to rollback, but don't override the original error
			return fmt.Errorf("executing deletion query (%s): %w", query, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// GetUserFromContext extracts and validates the user information from the Fiber context.
func (s *UserService) GetUserFromContext(c *fiber.Ctx) (*dto.UserData, error) {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	username, ok := c.Locals("username").(string)
	if !ok {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username not found"})
	}

	return &dto.UserData{
		UserID:   userID,
		Username: username,
	}, nil
}

func fetchNewUser(tx *sqlx.Tx, userID int64) (*model.User, error) {
	var newUser model.User
	query := `SELECT UserID, Username, Email, Provider, ProviderID, Role, CreatedAt, UpdatedAt FROM Users WHERE UserID = ?`
	err := tx.QueryRowx(query, userID).StructScan(&newUser)
	if err != nil {
		return nil, fmt.Errorf("error fetching newly created user: %w", err)
	}
	return &newUser, nil
}
