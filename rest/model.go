package rest

import (
	"encoding/json"
	"time"
)

type SortType int

const (
	sortTypeAsc SortType = iota
	sortTypeDesc
	sortTypeNone
)

type SortFieldType int

const (
	sortFieldTypeID SortFieldType = iota
	sortFieldTypeNode
	sortFieldTypeFrom
	sortFieldTypeTo
	sortFieldTypeWeight
	sortFieldTypeDescription
	sortFieldTypeTag
	sortFieldTypeCompileTime
	sortFieldTypeLastSeen
	sortFieldTypeFirmware
)

type GetListRequest struct {
	Filter map[string]any `form:"filter"`
	Range  string         `form:"range"`
	Sort   string         `form:"sort"`
}

type CreateNodeRequest struct {
	ID    uint   `json:"id"`
	Tag   string `json:"tag"`
	InUse bool   `json:"in_use"`
}

type UpdateNodeRequest struct {
	Tag      string `json:"tag"`
	InUse    bool   `json:"in_use"`
	Firmware string `json:"firmware"`
	DevTag   string `json:"dev_tag"`
	Channel  int8   `json:"channel"`
	TxPower  int8   `json:"tx_power"`
}

type MeshNodeFirmware struct {
	Title string `json:"title"`
	Src   string `json:"src"`
}

type MeshNode struct {
	ID          uint               `json:"id"`
	Tag         string             `json:"tag"`
	InUse       bool               `json:"in_use"`
	IsLocal     bool               `json:"is_local"`
	FirmRev     string             `json:"firmrev"`
	CompileTime string             `json:"comptime"`
	LastSeen    string             `json:"last_seen"`
	Firmware    []MeshNodeFirmware `json:"firmware"`
	Progress    int                `json:"progress"`
	Path        string             `json:"path"`
	Revision    string             `json:"revision"`
	Error       string             `json:"error"`
	DevTag      string             `json:"dev_tag"`
	Channel     int8               `json:"channel"`
	TxPower     int8               `json:"tx_power"`
	Groups      int                `json:"groups"`
	Binded      int                `json:"binded"`
	Flags       int                `json:"flags"`

	compileTime time.Time
	lastSeen    time.Time
}

func (n MeshNode) Sort(other MeshNode, sortType SortType, sortBy SortFieldType) bool {
	switch sortType {
	case sortTypeAsc:
		switch sortBy {
		case sortFieldTypeID:
			return n.ID < other.ID
		case sortFieldTypeNode:
			return n.ID < other.ID
		case sortFieldTypeTag:
			return n.Tag < other.Tag
		case sortFieldTypeFirmware:
			return n.FirmRev < other.FirmRev
		case sortFieldTypeCompileTime:
			return n.compileTime.Before(other.compileTime)
		case sortFieldTypeLastSeen:
			return n.lastSeen.Before(other.lastSeen)
		}
		return n.ID < other.ID
	case sortTypeDesc:
		switch sortBy {
		case sortFieldTypeID:
			return n.ID > other.ID
		case sortFieldTypeNode:
			return n.ID > other.ID
		case sortFieldTypeTag:
			return n.Tag > other.Tag
		case sortFieldTypeFirmware:
			return n.FirmRev > other.FirmRev
		case sortFieldTypeCompileTime:
			return n.compileTime.After(other.compileTime)
		case sortFieldTypeLastSeen:
			return n.lastSeen.After(other.lastSeen)
		}
		return n.ID > other.ID
	}

	return false
}

type CreateLinkRequest struct {
	From        int64   `json:"from"`
	To          int64   `json:"to"`
	Weight      float32 `json:"weight"`
	Description string  `json:"description"`
}

type UpdateLinkRequest struct {
	ID          uint    `json:"id"`
	Weight      float32 `json:"weight"`
	Description string  `json:"description"`
}

type MeshLink struct {
	ID          uint    `json:"id"`
	From        int64   `json:"from"`
	To          int64   `json:"to"`
	Weight      float32 `json:"weight"`
	Description string  `json:"description"`
}

func (l MeshLink) Sort(other MeshLink, sortType SortType, sortBy SortFieldType) bool {
	switch sortType {
	case sortTypeAsc:
		switch sortBy {
		case sortFieldTypeID:
			return l.ID < other.ID
		case sortFieldTypeNode:
			return l.ID < other.ID
		case sortFieldTypeFrom:
			return l.From < other.From
		case sortFieldTypeTo:
			return l.To < other.To
		case sortFieldTypeWeight:
			return l.Weight < other.Weight
		case sortFieldTypeDescription:
			return l.Description < other.Description
		}
		return l.ID < other.ID
	case sortTypeDesc:
		switch sortBy {
		case sortFieldTypeID:
			return l.ID > other.ID
		case sortFieldTypeNode:
			return l.ID > other.ID
		case sortFieldTypeFrom:
			return l.From > other.From
		case sortFieldTypeTo:
			return l.To > other.To
		case sortFieldTypeWeight:
			return l.Weight > other.Weight
		case sortFieldTypeDescription:
			return l.Description > other.Description
		}
		return l.ID > other.ID
	}
	return false
}

type CtrlDiscoveryRequest struct {
	Mode string `json:"mode"`
}

type MeshNeighbor struct {
	ID      uint    `json:"id"`
	Node    string  `json:"node"`
	Address string  `json:"address"`
	Current float32 `json:"current"`
	Next    float32 `json:"next"`
	Delta   float32 `json:"delta"`
}

type MeshDiscoveryState struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CurrentId string `json:"current_id"`
	Repeat    int    `json:"repeat"`
}

type MeshFirmware struct {
	ID       int64  `json:"id"`
	Status   string `json:"status"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Progress int    `json:"progress"`
}

type UpdateFirmwareRequest struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
}

var acceptFilters = map[string]struct{}{
	"from": {},
	"to":   {},
	"any":  {},
}

type EsphomeServer struct {
	ID      uint   `json:"id"`
	Address string `json:"address"`
	Clients int    `json:"clients"`
}

type EsphomeClient struct {
	ID   uint   `json:"id"`
	Node string `json:"Nodfr"`

	Address  string `json:"address"`
	Tag      string `json:"tag"`
	Active   bool   `json:"active"`
	Handle   int    `json:"handle"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
	Duration string `json:"duration"`
	Started  string `json:"started"`
}

type GetListParams struct {
	Filter        map[string]interface{}
	Limit, Offset int
	SortBy        SortFieldType
	SortType      SortType
}

func parseSortType(s string) SortType {
	switch s {
	case "ASC":
		return sortTypeAsc
	case "DESC":
		return sortTypeDesc
	}
	return sortTypeAsc
}

func parseSortFieldType(s string) SortFieldType {
	switch s {
	case "id":
		return sortFieldTypeID
	case "Node":
		return sortFieldTypeNode
	case "from":
		return sortFieldTypeFrom
	case "to":
		return sortFieldTypeTo
	case "weight":
		return sortFieldTypeWeight
	case "tag":
		return sortFieldTypeTag
	case "comptime":
		return sortFieldTypeCompileTime
	case "firmrev":
		return sortFieldTypeFirmware
	}
	return sortFieldTypeID
}

func (r GetListRequest) toGetListParams() GetListParams {
	acceptFiltersParam := make(map[string]any)
	for k, v := range r.Filter {
		if _, ok := acceptFilters[k]; ok {
			acceptFiltersParam[k] = v
		}
	}

	p := GetListParams{
		Filter:   acceptFiltersParam,
		Limit:    10,
		Offset:   0,
		SortBy:   sortFieldTypeID,
		SortType: sortTypeAsc,
	}

	if r.Range != "" {
		var queryRange []int
		_ = json.Unmarshal([]byte(r.Range), &queryRange)
		if len(queryRange) == 2 {
			p.Offset, p.Limit = queryRange[0], queryRange[1]
		}
	}

	if r.Sort != "" {
		var querySort []string
		_ = json.Unmarshal([]byte(r.Sort), &querySort)
		if len(querySort) == 2 {
			p.SortBy, p.SortType = parseSortFieldType(querySort[0]), parseSortType(querySort[1])
		}
	}

	return p
}

func formatTimeForJson(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
