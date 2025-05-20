package explainer

import (
	"database/sql"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
)

// AddIssueDetectionFeedback add a new feedback in the detection_feedback table (or update it if the user already posted a feedback)
// Moreover, it updates the issue average rating for convenience
func AddIssueDetectionFeedback(dbClient *sqlx.DB, issue model.Issue, user users.User, rating int) error {
	// FIXME: Alter detection feedback table to allow userID uuid.UUID instead of int64
	// tx, err := dbClient.Beginx()
	// if err != nil {
	// 	return err
	// }

	// err = persistDetectionFeedback(tx, issue.ID, user.ID, rating)
	// if err != nil {
	// 	return err
	// }

	// avg, err := calculateDetectionRatingAverage(tx, issue.ID)
	// if err != nil {
	// 	return err
	// }

	// err = updateIssueDetectionFeedbackAvg(tx, issue.ID, avg)
	// if err != nil {
	// 	return err
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	zap.L().Warn("AddIssueDetectionFeedback is not implemented")

	return nil
}

func persistDetectionFeedback(tx *sqlx.Tx, issueID int64, userID int64, rating int) error {
	query := `
		INSERT into issue_detection_feedback_v3 (issue_id, user_id, date, rating) 
			values (:issue_id, :user_id, :feedback_date, :rating) 
		ON CONFLICT ON CONSTRAINT unq_issueid_userid 
		DO UPDATE SET rating = :rating, date = :feedback_date 
		WHERE issue_detection_feedback_v3.issue_id = :issue_id AND issue_detection_feedback_v3.user_id = :user_id`
	params := map[string]interface{}{
		"issue_id":      issueID,
		"user_id":       userID,
		"feedback_date": time.Now().UTC(),
		"rating":        rating,
	}

	res, err := tx.NamedExec(query, params)
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}

	return nil
}

func calculateDetectionRatingAverage(tx *sqlx.Tx, issueID int64) (float64, error) {
	query := `select avg(rating) from issue_detection_feedback_v3 where issue_id = :issue_id;`
	params := map[string]interface{}{
		"issue_id": issueID,
	}
	rows, err := tx.NamedQuery(query, params)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if !rows.Next() {
		return -1, nil
	}
	var avgRating sql.NullFloat64
	err = rows.Scan(&avgRating)
	if err != nil {
		return -1, err
	}
	if avgRating.Valid {
		return avgRating.Float64, nil
	}
	return -1, nil
}

func updateIssueDetectionFeedbackAvg(tx *sqlx.Tx, issueID int64, avgRating float64) error {
	query := `update issues_v1 set detection_rating_avg = :rating_avg where id = :issue_id`
	params := map[string]interface{}{
		"rating_avg": avgRating,
		"issue_id":   issueID,
	}
	res, err := tx.NamedExec(query, params)
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}
