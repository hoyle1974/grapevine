package services

type AccountId string

func NilAccountId() AccountId {
	return NewAccountId("")
}

func (a AccountId) String() string {
	return string(a)
}

func NewAccountId(id string) AccountId {
	return AccountId(id)
}

func AccountIdsToStrings(accountIds []AccountId) []string {
	out := make([]string, len(accountIds))
	for _, accountId := range accountIds {
		out = append(out, accountId.String())
	}

	return out
}

func StringsToAccountIds(accountIds []string) []AccountId {
	out := make([]AccountId, len(accountIds))
	for _, accountId := range accountIds {
		out = append(out, NewAccountId(accountId))
	}

	return out
}
