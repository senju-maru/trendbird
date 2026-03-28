package converter

import "time"

// timeToOptionalString は *time.Time を RFC3339 形式の *string に変換する。nil の場合は nil を返す。
func timeToOptionalString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
