package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// StringArray is a custom type for PostgreSQL TEXT[] columns.
type StringArray []string

// Scan implements the sql.Scanner interface for TEXT[].
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = StringArray{}
		return nil
	}
	s, ok := src.(string)
	if !ok {
		if b, ok := src.([]byte); ok {
			s = string(b)
		} else {
			return fmt.Errorf("StringArray.Scan: unsupported type %T", src)
		}
	}
	// PostgreSQL TEXT[] format: {val1,val2,val3} or {}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		*a = StringArray{}
		return nil
	}
	parts := strings.Split(s, ",")
	result := make(StringArray, len(parts))
	for i, p := range parts {
		// Remove surrounding quotes if present
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"")
		result[i] = p
	}
	*a = result
	return nil
}

// Value implements the driver.Valuer interface for TEXT[].
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	quoted := make([]string, len(a))
	for i, s := range a {
		// Escape double quotes and backslashes
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		quoted[i] = fmt.Sprintf(`"%s"`, s)
	}
	return fmt.Sprintf("{%s}", strings.Join(quoted, ",")), nil
}

// AutoDMRule is the GORM model for the auto_dm_rules table.
type AutoDMRule struct {
	ID                 string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string      `gorm:"type:uuid;not null;index:idx_auto_dm_rules_user_id"`
	Enabled            bool        `gorm:"not null;default:false"`
	TriggerKeywords    StringArray `gorm:"type:text[];not null;default:'{}'"`
	TemplateMessage    string      `gorm:"type:text;not null;default:''"`
	LastCheckedReplyID *string     `gorm:"type:varchar(64)"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (AutoDMRule) TableName() string {
	return "auto_dm_rules"
}
