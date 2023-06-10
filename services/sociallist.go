package services

import (
	"fmt"

	"github.com/hoyle1974/grapevine/common"
)

type SocialListType string

const (
	SocialListType_BLOCKED   SocialListType = "BLOCKED"
	SocialListType_FOLLOWS   SocialListType = "FOLLOWS"
	SocialListType_FOLLOWING SocialListType = "FOLLOWING"
)

func GetSocialList(appCtx AppCtx, owner common.AccountId, listType SocialListType) ([]common.AccountId, error) {
	log := appCtx.Log("GetSocialList")
	log.Printf("Received: %v", owner)

	stmt := `select entity_id from "lists" where "owner_id" = $1 and "list_type" = $2`
	if listType == SocialListType_FOLLOWING {
		stmt = `select owner_id from "lists" where "entity_id" = $1 and "list_type" = $2`
		listType = SocialListType_FOLLOWS
	}
	rows, err := appCtx.db.Query(stmt, owner.String(), listType)
	if err != nil {
		return []common.AccountId{}, err
	}

	entities := make([]common.AccountId, 0)
	defer rows.Close()
	for rows.Next() {
		var entity_id string
		rows.Scan(&entity_id)
		entities = append(entities, common.NewAccountId(entity_id))
	}

	return entities, nil
}

func AddToSocialList(appCtx AppCtx, owner common.AccountId, listType SocialListType, idToAdd common.AccountId) error {
	log := appCtx.Log("AddToSocialList")
	log.Printf("Received: %v/%v/%v", owner, listType, idToAdd)
	if listType == SocialListType_FOLLOWING {
		return fmt.Errorf("SocialListType_FOLLOWING is a virtual type and can not be inserted")
	}

	stmt := `insert into "lists"("id", "list_type","owner_id","entity_id") values(gen_random_uuid(),$1, $2,$3)`
	row := appCtx.db.QueryRow(stmt, listType, owner.String(), idToAdd.String())
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

func RemoveFromSocialList(appCtx AppCtx, owner common.AccountId, listType string, idToRemove common.AccountId) error {
	log := appCtx.Log("RemoveFromSocialList")
	log.Printf("Received: %v/%v/%v", owner, listType, idToRemove)

	stmt := `delete from "lists" where "owner_id" = ? and "list_type" = ? and "entity_id" = ?)`
	result, err := appCtx.db.Exec(stmt, owner.String(), listType, idToRemove.String())
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("Removed %d rows\n", rows)

	return nil
}
