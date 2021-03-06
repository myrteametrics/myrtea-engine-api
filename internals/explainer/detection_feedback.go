package explainer

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
)

// AddIssueDetectionFeedback add a new feedback in the detection_feedback table (or update it if the user already posted a feedback)
// Moreover, it updates the issue average rating for convenience
func AddIssueDetectionFeedback(dbClient *sqlx.DB, issueID int64, userID int64, rating int, groups []int64) error {
	_, found, err := issues.R().Get(issueID, groups)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Issue with id %d not found", issueID)
	}

	tx, err := dbClient.Beginx()
	if err != nil {
		return err
	}

	err = persistDetectionFeedback(tx, issueID, userID, rating)
	if err != nil {
		return err
	}

	avg, err := calculateDetectionRatingAverage(tx, issueID)
	if err != nil {
		return err
	}

	err = updateIssueDetectionFeedbackAvg(tx, issueID, avg)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

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
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
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
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}
