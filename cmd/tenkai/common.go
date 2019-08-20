package main

func (appContext *appContext) hasAccess(email string, envID int) (bool, error) {
	result := false
	environments, err := appContext.database.GetAllEnvironments(email)
	if err != nil {
		return false, err
	}
	for _, e := range environments {
		if e.ID == uint(envID) {
			result = true
			break
		}
	}
	return result, nil
}
