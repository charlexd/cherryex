package cherryFacade

type (
	SID = string // session unique id
	UID = int64  // user unique id
)

func ToUID(from int32) UID {
	return UID(from)
}
