package member

type Role string

const (
	RoleViewer   Role = "VIEWER"
	RoleDesigner Role = "DESIGNER"
	RoleAdmin    Role = "ADMIN"
)

func (r Role) AtLeast(minimum Role) bool {
	order := map[Role]int{
		RoleViewer:   0,
		RoleDesigner: 1,
		RoleAdmin:    2,
	}
	return order[r] >= order[minimum]
}

func (r Role) IsValid() bool {
	switch r {
	case RoleViewer, RoleDesigner, RoleAdmin:
		return true
	}
	return false
}
