package api

import (
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/vmkteam/zenrpc/v2"
)

type PublicService struct {
	zenrpc.Service
}

func NewPublicService() *PublicService {
	return &PublicService{}
}

// Gets all supported go-pg versions
//zenrpc:return	list of versions
func (s *PublicService) GoPGVersions() []int {
	return []int{
		mfd.GoPG8,
		mfd.GoPG9,
		mfd.GoPG10,
	}
}

// Gets all available entity modes
//zenrpc:return	list of modes
func (s *PublicService) Modes() []string {
	return []string{
		mfd.ModeFull,
		mfd.ModeReadOnlyWithTemplates,
		mfd.ModeReadOnly,
		mfd.ModeNone,
	}
}

// Gets all available search types
//zenrpc:return	list of search types
func (s *PublicService) SearchTypes() []string {
	return []string{
		mfd.SearchEquals,
		mfd.SearchNotEquals,
		mfd.SearchNull,
		mfd.SearchNotNull,
		mfd.SearchGE,
		mfd.SearchLE,
		mfd.SearchG,
		mfd.SearchL,
		mfd.SearchLeftLike,
		mfd.SearchLeftILike,
		mfd.SearchRightLike,
		mfd.SearchRightILike,
		mfd.SearchLike,
		mfd.SearchILike,
		mfd.SearchArray,
		mfd.SearchNotArray,
		mfd.SearchTypeArrayContains,
		mfd.SearchTypeArrayNotContains,
		mfd.SearchTypeArrayContained,
		mfd.SearchTypeArrayIntersect,
		mfd.SearchTypeJsonbPath,
	}
}

// Gets std types
//zenrpc:return	list of types
func (s *PublicService) Types() []string {
	return []string{
		model.TypeInt,
		model.TypeInt32,
		model.TypeInt64,
		model.TypeFloat32,
		model.TypeFloat64,
		model.TypeString,
		model.TypeByteSlice,
		model.TypeBool,
		model.TypeTime,
		model.TypeDuration,
		model.TypeMapInterface,
		model.TypeMapString,
		model.TypeIP,
		model.TypeIPNet,
		model.TypeInterface,
	}
}

// Gets postgres types
//zenrpc:return	list of types
func (s *PublicService) DBTypes() []string {
	return []string{
		model.TypePGInt2,
		model.TypePGInt4,
		model.TypePGInt8,
		model.TypePGNumeric,
		model.TypePGFloat4,
		model.TypePGFloat8,
		model.TypePGText,
		model.TypePGVarchar,
		model.TypePGUuid,
		model.TypePGBpchar,
		model.TypePGBytea,
		model.TypePGBool,
		model.TypePGTimestamp,
		model.TypePGTimestamptz,
		model.TypePGDate,
		model.TypePGTime,
		model.TypePGTimetz,
		model.TypePGInterval,
		model.TypePGJSONB,
		model.TypePGJSON,
		model.TypePGHstore,
		model.TypePGInet,
		model.TypePGCidr,
		model.TypePGPoint,
	}
}

func (s *PublicService) Ping() string {
	return "Pong"
}

//go:generate zenrpc
