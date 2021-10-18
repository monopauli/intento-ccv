package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	//"cosmos/base/v1beta1/coin.proto"
)

var _ sdk.Msg = &MsgCreateItem{}

func NewMsgCreateItem(creator string, title string, description string, shippingcost int64, localpickup string, estimationcount int64, tags []string, condition int64, shippingregion []string, depositamount int64, initmsg []byte, photos []string) *MsgCreateItem {
	return &MsgCreateItem{

		Creator:         creator,
		Title:           title,
		Description:     description,
		Shippingcost:    shippingcost,
		Localpickup:     localpickup,
		Estimationcount: estimationcount,
		Tags:            tags,
		Condition:       condition,
		Shippingregion:  shippingregion,
		Depositamount:   depositamount,
		Initmsg:         initmsg,
		Photos:          photos,
	}
}

func (msg *MsgCreateItem) Route() string {
	return RouterKey
}

func (msg *MsgCreateItem) Type() string {
	return "CreateItem"
}

func (msg *MsgCreateItem) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgCreateItem) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateItem) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}

	if len(msg.Tags) > 5 || len(msg.Tags) < 1 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "tags invalid")
	}
	for _, tags := range msg.Tags {
		if len(tags) > 24 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "a tag was too long")
		}
	}

	if len(msg.Shippingregion) > 9 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Region list too long")
	}

	for _, region := range msg.Shippingregion {
		if len(region) > 2 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "A Region cannot be longer than 2")
		}
	}

	if msg.Shippingcost == 0 && msg.Localpickup == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Provide either shipping or localpickup")
	}

	if len(msg.Description) > 1000 {
		return sdkerrors.Wrap(sdkerrors.ErrMemoTooLarge, "description too long")
	}

	if len(msg.Localpickup) > 48 {
		return sdkerrors.Wrap(sdkerrors.ErrMemoTooLarge, "Local pickup location too long")
	}

	if msg.Condition > 5 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid item condition")
	}
	if msg.Estimationcount > 24 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid estimation count")
	}

	if len(msg.Photos) > 9 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "too many photos")
	}
	for _, photo := range msg.Photos {
		if len(photo) > 200 {
			return sdkerrors.Wrap(sdkerrors.ErrMemoTooLarge, "photo url too long")
		}
	}
	return nil
}

var _ sdk.Msg = &MsgCreateItem{}

func NewMsgDeleteItem(seller string, id uint64) *MsgDeleteItem {
	return &MsgDeleteItem{
		Id:     id,
		Seller: seller,
	}
}
func (msg *MsgDeleteItem) Route() string {
	return RouterKey
}

func (msg *MsgDeleteItem) Type() string {
	return "DeleteItem"
}

func (msg *MsgDeleteItem) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgDeleteItem) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteItem) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	return nil
}

func NewMsgRevealEstimation(creator string, itemid uint64, revealmsg []byte) *MsgRevealEstimation {
	return &MsgRevealEstimation{

		Creator:   creator,
		Itemid:    itemid,
		Revealmsg: revealmsg,
	}
}

func (msg *MsgRevealEstimation) Route() string {
	return RouterKey
}

func (msg *MsgRevealEstimation) Type() string {
	return "RevealEstimation"
}

func (msg *MsgRevealEstimation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRevealEstimation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRevealEstimation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateItem{}

func NewMsgItemTransferable(seller string, transferable bool, itemid uint64) *MsgItemTransferable {
	return &MsgItemTransferable{

		Seller:       seller,
		Transferable: transferable,
		Itemid:       itemid,
	}
}

func (msg *MsgItemTransferable) Route() string {
	return RouterKey
}

func (msg *MsgItemTransferable) Type() string {
	return "ItemTransferable"
}

func (msg *MsgItemTransferable) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgItemTransferable) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgItemTransferable) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateItem{}

func NewMsgItemShipping(seller string, tracking bool, itemid uint64) *MsgItemShipping {
	return &MsgItemShipping{

		Seller:   seller,
		Tracking: tracking,
		Itemid:   itemid,
	}
}

func (msg *MsgItemShipping) Route() string {
	return RouterKey
}

func (msg *MsgItemShipping) Type() string {
	return "ItemShipping"
}

func (msg *MsgItemShipping) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgItemShipping) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgItemShipping) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgCreateItem{}

func NewMsgItemResell(seller string, itemid uint64, shippingcost int64, discount int64, localpickup string, shippingregion []string, note string) *MsgItemResell {
	return &MsgItemResell{
		Seller:         seller,
		Itemid:         itemid,
		Shippingcost:   shippingcost,
		Discount:       discount,
		Localpickup:    localpickup,
		Shippingregion: shippingregion,
		Note:           note,
	}
}

func (msg *MsgItemResell) Route() string {
	return RouterKey
}

func (msg *MsgItemResell) Type() string {
	return "ItemResell"
}

func (msg *MsgItemResell) GetSigners() []sdk.AccAddress {
	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{seller}
}

func (msg *MsgItemResell) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgItemResell) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seller address (%s)", err)
	}
	if len(msg.Note) > 240 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "note too long")
	}
	if len(msg.Shippingregion) > 6 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Regions too long")
	}

	for _, region := range msg.Shippingregion {
		if len(region) > 2 {
			return sdkerrors.Wrap(sdkerrors.ErrMemoTooLarge, "Region too long")
		}
	}
	if msg.Shippingcost == 0 && msg.Localpickup == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Provide either shipping or localpickup")
	}

	if len(msg.Localpickup) > 25 {
		return sdkerrors.Wrap(sdkerrors.ErrMemoTooLarge, "Local pickup too long")
	}

	return nil
}

func NewMsgTokenizeItem(sender string, id uint64) *MsgTokenizeItem {
	return &MsgTokenizeItem{
		Id:     id,
		Sender: sender,
	}
}
func (msg *MsgTokenizeItem) Route() string {
	return RouterKey
}

func (msg *MsgTokenizeItem) Type() string {
	return "DeleteItem"
}

func (msg *MsgTokenizeItem) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgTokenizeItem) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTokenizeItem) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}
	return nil
}

func NewMsgUnTokenizeItem(sender string, id uint64) *MsgUnTokenizeItem {
	return &MsgUnTokenizeItem{
		Id:     id,
		Sender: sender,
	}
}
func (msg *MsgUnTokenizeItem) Route() string {
	return RouterKey
}

func (msg *MsgUnTokenizeItem) Type() string {
	return "DeleteItem"
}

func (msg *MsgUnTokenizeItem) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgUnTokenizeItem) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnTokenizeItem) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}
	return nil
}
