package groups

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

// PostgresRepository is a repository containing the user groups data based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

//Get search and returns an User Group from the repository by its id
func (r *PostgresRepository) Get(id int64) (Group, bool, error) {
	query := `SELECT id, name FROM user_groups_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return Group{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		group := Group{}
		err = rows.Scan(&group.ID, &group.Name)
		if err != nil {
			return Group{}, false, err
		}
		return group, true, nil
	}

	return Group{}, false, nil
}

// Create creates a new User Group in the repository
func (r *PostgresRepository) Create(group Group) (int64, error) {
	query := `INSERT INTO user_groups_v1 VALUES (DEFAULT, :name) RETURNING id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name": group.Name,
	})
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return -1, errors.New("Error creating Group " + err.Error())
	}
	return id, nil

}

// Update updates an User Group in the repository
func (r *PostgresRepository) Update(group Group) error {
	query := `UPDATE user_groups_v1 SET name = :name WHERE id = :id`
	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id":   group.ID,
		"name": group.Name,
	})
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Delete deletes an User Group in the repository
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM user_groups_v1 WHERE id = :id`
	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("No row deleted (or multiple row deleted) instead of 1 row")
	}
	return nil
}

// GetAll returns all User Groups in the repository
func (r *PostgresRepository) GetAll() (map[int64]Group, error) {
	query := `SELECT id, name FROM user_groups_v1`
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, errors.New("Couldn't retrieve the User Groups " + err.Error())
	}
	defer rows.Close()

	groups := make(map[int64]Group, 0)
	for rows.Next() {
		group := Group{}
		err := rows.Scan(&group.ID, &group.Name)
		if err != nil {
			return nil, errors.New("Couldn't scan the retrieved data: " + err.Error())
		}
		groups[group.ID] = group
	}
	return groups, nil
}

// CreateMembership creates a new User Membership in the repository
func (r *PostgresRepository) CreateMembership(membership Membership) error {
	query := `INSERT INTO user_memberships_v1 VALUES (:user_id, :group_id, :role)`
	params := map[string]interface{}{
		"user_id":  membership.UserID,
		"group_id": membership.GroupID,
		"role":     membership.Role,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't create the User Membership " + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

//GetMembership search and returns an User membership
func (r *PostgresRepository) GetMembership(userID int64, groupID int64) (Membership, bool, error) {

	query := `SELECT role FROM user_memberships_v1 WHERE user_id = :user_id AND group_id = :group_id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"user_id":  userID,
		"group_id": groupID,
	})
	if err != nil {
		return Membership{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		membership := Membership{UserID: userID, GroupID: groupID}
		err = rows.Scan(&membership.Role)
		if err != nil {
			return Membership{}, false, err
		}
		return membership, true, nil
	}

	return Membership{}, false, nil
}

// UpdateMembership updates an User Membership in the repository
func (r *PostgresRepository) UpdateMembership(membership Membership) error {
	query := `UPDATE user_memberships_v1 SET role = :role WHERE user_id = :user_id AND group_id = :group_id`
	params := map[string]interface{}{
		"user_id":  membership.UserID,
		"group_id": membership.GroupID,
		"role":     membership.Role,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't update the User Membership:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row updated (or multiple row updated) instead of 1 row")
	}
	return nil
}

// DeleteMembership deletes an User Membership in the repository
func (r *PostgresRepository) DeleteMembership(userID int64, groupID int64) error {
	query := `DELETE FROM user_memberships_v1 WHERE user_id = :user_id AND group_id = :group_id`
	params := map[string]interface{}{
		"user_id":  userID,
		"group_id": groupID,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't delete the User Membership:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row deleted (or multiple row deleted) instead of 1 row")
	}
	return nil
}

// GetGroupsOfUser returns all User Groups of a User in the repository
func (r *PostgresRepository) GetGroupsOfUser(userID int64) ([]GroupOfUser, error) {
	query := `SELECT user_groups_v1.id, user_groups_v1.name, user_memberships_v1.role AS role_in_group 
		FROM user_memberships_v1 INNER JOIN user_groups_v1 ON user_memberships_v1.group_id = user_groups_v1.id 
		WHERE user_memberships_v1.user_id = :user_id`
	params := map[string]interface{}{
		"user_id": userID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, errors.New("Couldn't retrieve the Groups of User with id " + string(userID) + " : " + err.Error())
	}
	defer rows.Close()

	groups := make([]GroupOfUser, 0)
	for rows.Next() {
		group := GroupOfUser{}
		err := rows.Scan(&group.ID, &group.Name, &group.UserRole)
		if err != nil {
			return nil, errors.New("Couldn't scan the retrieved data: " + err.Error())
		}
		groups = append(groups, group)
	}
	return groups, nil
}
