package api

import (
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/vmkteam/zenrpc/v2"
)

//go:generate zenrpc

type PublicService struct {
	zenrpc.Service
}

func NewPublicService() *PublicService {
	return &PublicService{}
}

// GoPGVersions returns all supported go-pg versions.
//
//zenrpc:return	list of versions
func (s PublicService) GoPGVersions() []int {
	return []int{
		mfd.GoPG8,
		mfd.GoPG9,
		mfd.GoPG10,
	}
}

// Modes returns all available entity modes.
//
//zenrpc:return	list of modes
func (s PublicService) Modes() []string {
	return []string{
		mfd.ModeFull,
		mfd.ModeReadOnlyWithTemplates,
		mfd.ModeReadOnly,
		mfd.ModeNone,
	}
}

// SearchTypes returns all available search types.
//
//zenrpc:return	list of search types
func (s PublicService) SearchTypes() []string {
	return []string{
		mfd.SearchEquals.String(),
		mfd.SearchNotEquals.String(),
		mfd.SearchNull.String(),
		mfd.SearchNotNull.String(),
		mfd.SearchGE.String(),
		mfd.SearchLE.String(),
		mfd.SearchG.String(),
		mfd.SearchL.String(),
		mfd.SearchLeftLike.String(),
		mfd.SearchLeftILike.String(),
		mfd.SearchRightLike.String(),
		mfd.SearchRightILike.String(),
		mfd.SearchLike.String(),
		mfd.SearchILike.String(),
		mfd.SearchArray.String(),
		mfd.SearchNotArray.String(),
		mfd.SearchTypeArrayContains.String(),
		mfd.SearchTypeArrayNotContains.String(),
		mfd.SearchTypeArrayContained.String(),
		mfd.SearchTypeArrayIntersect.String(),
		mfd.SearchTypeJsonbPath.String(),
	}
}

// Types returns list of types.
//
//zenrpc:return	list of types
func (s PublicService) Types() []string {
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

// DBTypes returns postgres types.
//
//zenrpc:return	list of types
func (s PublicService) DBTypes() []string {
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

func (s PublicService) HTMLTypes() []string {
	return []string{
		mfd.TypeHTMLNone,
		mfd.TypeHTMLInput,
		mfd.TypeHTMLText,
		mfd.TypeHTMLPassword,
		mfd.TypeHTMLEditor,
		mfd.TypeHTMLCheckbox,
		mfd.TypeHTMLDateTime,
		mfd.TypeHTMLDate,
		mfd.TypeHTMLTime,
		mfd.TypeHTMLSelect,
		mfd.TypeHTMLFile,
		mfd.TypeHTMLImage,
	}
}

func (s PublicService) Ping() string {
	return "Pong"
}
