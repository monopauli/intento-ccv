package keeper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	sdkmath "cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
	"github.com/trstlabs/intento/x/intent/types"
	icqtypes "github.com/trstlabs/intento/x/interchainquery/types"
)

// CompareResponseValue compares the value of a response key based on the ResponseComparison
func (k Keeper) CompareResponseValue(ctx sdk.Context, actionID uint64, responses []*cdctypes.Any, comparison types.ResponseComparison, queryCallback *icqtypes.Query) (bool, error) {
	if comparison.ResponseKey == "" {
		return true, nil
	}

	if len(responses) <= int(comparison.ResponseIndex) {
		return false, fmt.Errorf("not enough responses to compare, number of responses: %v", len(responses))
	}
	var resp interface{}
	err := k.cdc.UnpackAny((responses)[comparison.ResponseIndex], &resp)
	if err != nil {
		return false, fmt.Errorf("error unpacking: %v", err)
	}
	value, err := ParseResponseValue(resp, comparison.ResponseKey, comparison.ValueType)
	if err != nil {
		return false, fmt.Errorf("error parsing value: %v", err)
	}
	operand, err := ParseOperand(comparison.ComparisonOperand, comparison.ValueType)
	if err != nil {
		return false, fmt.Errorf("error parsing operand: %v", err)
	}
	switch comparison.ComparisonOperator {
	case types.ComparisonOperator_EQUAL:
		return reflect.DeepEqual(value, operand), nil
	case types.ComparisonOperator_NOT_EQUAL:
		return !reflect.DeepEqual(value, operand), nil
	case types.ComparisonOperator_CONTAINS:
		return contains(value, operand), nil
	case types.ComparisonOperator_NOT_CONTAINS:
		return !contains(value, operand), nil
	case types.ComparisonOperator_SMALLER_THAN:
		return compareNumbers(value, operand, func(a, b int) bool { return a < b })
	case types.ComparisonOperator_LARGER_THAN:
		return compareNumbers(value, operand, func(a, b int) bool { return a > b })
	case types.ComparisonOperator_GREATER_EQUAL:
		return compareNumbers(value, operand, func(a, b int) bool { return a >= b })
	case types.ComparisonOperator_LESS_EQUAL:
		return compareNumbers(value, operand, func(a, b int) bool { return a <= b })
	case types.ComparisonOperator_STARTS_WITH:
		return strings.HasPrefix(value.(string), operand.(string)), nil
	case types.ComparisonOperator_ENDS_WITH:
		return strings.HasSuffix(value.(string), operand.(string)), nil
	default:
		return false, fmt.Errorf("unsupported comparison operator: %v", comparison.ComparisonOperator)
	}
}

// UseResponseValue replaces the value in a message with the value from a response
func (k Keeper) UseResponseValue(ctx sdk.Context, actionID uint64, msgs *[]*cdctypes.Any, conditions *types.ExecutionConditions, queryCallback *icqtypes.Query) error {
	if conditions == nil || conditions.UseResponseValue == nil || conditions.UseResponseValue.ResponseKey == "" {
		return nil
	}

	useResp := conditions.UseResponseValue
	if useResp.ActionID != 0 {
		actionID = useResp.ActionID

	}
	history, err := k.GetActionHistory(ctx, actionID)
	if err != nil {
		return err
	}
	if len(history) == 0 {
		return nil
	}
	var valueFromResponse interface{}
	if queryCallback == nil {
		responsesAnys := history[len(history)-1].MsgResponses
		if len(responsesAnys) == 0 {
			return nil
		}
		if int(useResp.ResponseIndex+1) < len(responsesAnys) {
			return fmt.Errorf("response index out of range")
		}

		protoMsg, err := k.interfaceRegistry.Resolve(responsesAnys[useResp.ResponseIndex].TypeUrl)
		if err != nil {
			return fmt.Errorf("failed to resolve type URL %s: %w", responsesAnys[useResp.ResponseIndex].TypeUrl, err)
		}

		err = proto.Unmarshal(responsesAnys[useResp.ResponseIndex].Value, protoMsg)
		if err != nil {
			return err
		}

		k.Logger(ctx).Debug("use response value", "protoMsg", protoMsg.String(), "TypeUrl", responsesAnys[useResp.ResponseIndex].TypeUrl)
		valueFromResponse, err = ParseResponseValue(protoMsg, useResp.ResponseKey, useResp.ValueType)
		if err != nil {
			return err
		}
	} else {
		valueFromResponse, err = ParseResponseValue(queryCallback.CallbackData, useResp.ResponseKey, useResp.ValueType)
		if err != nil {
			return err
		}
	}
	var msgToInterface sdk.Msg
	//var msgToInterface interface{}
	msgAny := (*msgs)[useResp.MsgsIndex]
	if msgAny.TypeUrl == sdk.MsgTypeURL(&authztypes.MsgExec{}) {
		msgExec := &authztypes.MsgExec{}
		if err := proto.Unmarshal(msgAny.Value, msgExec); err != nil {
			return err
		}
		msgAny = msgExec.Msgs[0]
	}

	err = k.cdc.UnpackAny(msgAny, &msgToInterface)
	if err != nil {
		return err
	}

	//k.Logger(ctx).Debug("use response value", "interface", msgToInterface, "valueFromResponse", valueFromResponse)

	msgProto, ok := msgToInterface.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msgToInterface)
	}
	msgTo := reflect.ValueOf(msgProto)

	// If the value is a pointer, get the element it points to
	if msgTo.Kind() == reflect.Ptr {
		msgTo = msgTo.Elem()
	}

	// Ensure we're dealing with a struct
	if msgTo.Kind() != reflect.Struct {
		// k.Logger(ctx).Debug("use response value", "msgTo.Kind", msgTo.Kind())
		return fmt.Errorf("expected a struct, got %v", msgTo.Kind())
	}

	fieldToReplace, err := traverseFields(msgToInterface, useResp.MsgKey)
	if err != nil {
		return err
	}
	k.Logger(ctx).Debug("use response value", "fieldToReplace", fieldToReplace)

	// Set the new value
	if fieldToReplace.CanSet() {
		fieldToReplace.Set(reflect.ValueOf(valueFromResponse))
	} else {
		return fmt.Errorf("field %s cannot be set", fieldToReplace)
	}

	newMsgAny, err := cdctypes.NewAnyWithValue(msgToInterface)
	if err != nil {
		// k.Logger(ctx).Debug("use response value", "err", err, "newMsgAny", newMsgAny)
		return err
	}

	if (*msgs)[useResp.MsgsIndex].TypeUrl == sdk.MsgTypeURL(&authztypes.MsgExec{}) {
		msgExec := &authztypes.MsgExec{}
		if err := proto.Unmarshal((*msgs)[useResp.MsgsIndex].Value, msgExec); err != nil {
			return err
		}
		msgExec.Msgs[0] = newMsgAny
		k.Logger(ctx).Debug("use response value", "msgExec", msgExec.String())
		msgExecAnys, err := types.PackTxMsgAnys([]sdk.Msg{msgExec})
		if err != nil {
			return err
		}
		newMsgAny = msgExecAnys[0]
	}
	k.Logger(ctx).Debug("use response value", "newMsgAny", newMsgAny.TypeUrl)

	(*msgs)[useResp.MsgsIndex] = newMsgAny
	return nil
}

// ParseResponseValue retrieves and parses the value of a response key to the specified response type
func ParseResponseValue(response interface{}, responseKey, responseType string) (interface{}, error) {
	// val := reflect.ValueOf(response)

	// // If the value is a pointer, get the element it points to
	// if val.Kind() == reflect.Ptr {
	// 	val = val.Elem()
	// }

	// // Ensure we're dealing with a struct
	// if val.Kind() != reflect.Struct {
	// 	return nil, fmt.Errorf("expected a struct, got %v", val.Kind())
	// }

	// field := val.FieldByName(responseKey)
	// if !field.IsValid() {
	// 	return nil, fmt.Errorf("field %s not found", responseKey)
	// }

	val, err := traverseFields(response, responseKey)
	if err != nil {
		// if responseKey == ""{
		// 	val =
		// }
		return nil, err
	}

	switch responseType {
	case "string":
		if val.Kind() == reflect.String {
			return val.String(), nil
		}
	case "sdk.Coin":
		if val.Kind() == reflect.Slice && val.Type().Elem().Name() == "Coin" {
			coins := val.Interface().(sdk.Coins)
			return coins[0], nil
		}
		if val.Kind() == reflect.Struct {
			amountField := val.FieldByName("Amount")
			denomField := val.FieldByName("Denom")
			if amountField.IsValid() && denomField.IsValid() && amountField.Type() == reflect.TypeOf(sdkmath.Int{}) && denomField.Kind() == reflect.String {
				amount := amountField.Interface().(sdkmath.Int)
				return sdk.Coin{
					Amount: amount,
					Denom:  denomField.String(),
				}, nil
			}
		}
	case "sdk.Coins":
		if val.Kind() == reflect.Slice && val.Type().Elem().Name() == "Coin" {
			coins := val.Interface().(sdk.Coins)
			return coins, nil
		}
	case "sdk.Int":
		if val.Kind() == reflect.Struct && val.Type().Name() == "Int" {
			return val.Interface().(sdk.Int), nil
		}
	case "[]string":
		if val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.String {
			return val.Interface().([]string), nil
		}
	case "[]sdk.Int":
		if val.Kind() == reflect.Slice && val.Type().Elem().Name() == "Int" {
			return val.Interface().([]sdkmath.Int), nil
		}
		// case "[]sdk.Coin":
		// 	if val.Kind() == reflect.Slice && val.Type().Elem().Name() == "Coin" {
		// 		return val.Interface().([]sdk.Coin), nil
		// 	}
	}

	return nil, fmt.Errorf("field %s could not be parsed as %s", responseKey, responseType)
}

// ParseOperand parses the operand based on the response type
func ParseOperand(operand string, responseType string) (interface{}, error) {
	switch responseType {
	case "string":
		return operand, nil
	case "sdk.Coin":
		coin, err := sdk.ParseCoinNormalized(operand)
		return coin, err
	case "sdk.Coins":
		coins, err := sdk.ParseCoinsNormalized(operand)
		return coins, err
	case "sdk.Int":
		var sdkInt sdkmath.Int
		sdkInt, ok := sdk.NewIntFromString(operand)
		if !ok {
			return nil, fmt.Errorf("unsupported int operand")
		}
		return sdkInt, nil
	case "[]string":
		var strArr []string
		err := json.Unmarshal([]byte(operand), &strArr)
		return strArr, err
	case "[]sdk.Int":
		var intArr []sdkmath.Int
		err := json.Unmarshal([]byte(operand), &intArr)
		return intArr, err
		// case "[]sdk.Coin":
		// 	coinArr, err := sdk.ParseCoinsNormalized(operand)
		// 	return coinArr, err
	}
	return nil, fmt.Errorf("unsupported operand type: %s", responseType)
}

// contains checks if a value contains an operand
func contains(value, operand interface{}) bool {
	val := reflect.ValueOf(value)
	operandVal := reflect.ValueOf(operand)

	switch val.Kind() {
	case reflect.String:
		return strings.Contains(val.String(), operandVal.String())
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(val.Index(i).Interface(), operandVal.Interface()) {
				return true
			}
		}
	}
	// Custom case for sdk.Coins
	if coins, ok := value.(sdk.Coins); ok {
		if coin, ok := operand.(sdk.Coins); ok {
			_, notOk := coins.SafeSub(coin...)
			return !notOk
		}
	}
	return false
}

// compareNumbers compares two numeric values based on a provided comparison function
func compareNumbers(value, operand interface{}, compareFunc func(int, int) bool) (bool, error) {
	val := reflect.ValueOf(value)
	operandVal := reflect.ValueOf(operand)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareFunc(int(val.Int()), int(operandVal.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compareFunc(int(val.Uint()), int(operandVal.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return compareFunc(int(val.Float()), int(operandVal.Float())), nil
	case reflect.Struct:
		if val.Type().Name() == "Int" {
			return compareFunc(int(val.MethodByName("Int64").Call(nil)[0].Int()), int(operandVal.MethodByName("Int64").Call(nil)[0].Int())), nil
		}
	}
	return false, fmt.Errorf("unsupported numeric type: %v", val.Kind())
}

// TraverseFields traverses the nested fields of a struct or slice based on the provided keys
func traverseFields(msgInterface interface{}, inputKey string) (reflect.Value, error) {
	keys := strings.Split(inputKey, ".")
	val := reflect.ValueOf(msgInterface)
	for _, key := range keys {
		// Handle slices
		if val.Kind() == reflect.Slice {
			index, err := parseIndex(key)
			if err != nil {
				return reflect.Value{}, err
			}
			if index >= val.Len() {
				return reflect.Value{}, fmt.Errorf("index %d out of bounds for slice", index)
			}
			val = val.Index(index)
		} else {
			// If the value is a pointer, get the element it points to
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}

			// Ensure we're dealing with a struct
			if val.Kind() != reflect.Struct {
				return reflect.Value{}, fmt.Errorf("expected a struct, got %v", val.Kind())
			}

			val = val.FieldByName(key)
			if !val.IsValid() {
				return reflect.Value{}, fmt.Errorf("field %s not found", key)
			}
		}
	}
	return val, nil
}

// parseIndex parses a string index for slices
func parseIndex(key string) (int, error) {
	var index int
	_, err := fmt.Sscanf(key, "[%d]", &index)
	if err != nil {
		return -1, fmt.Errorf("invalid slice index: %s", key)
	}
	return index, nil
}
