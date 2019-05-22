package hive

import (
	"errors"
	"github.com/taozle/go-hive-driver/thriftlib"
	"strconv"
)

func convertColumnValues(column *thriftlib.TColumnDesc, value *thriftlib.TColumn) (interface{}, error) {
	if column.GetTypeDesc().GetTypes()[0].PrimitiveEntry == nil {
		// string
		if !value.IsSetStringVal() {
			return nil, errors.New("hive: column is not primitive entry and value is not string")
		}

		return value.GetStringVal().GetValues(), nil
	}

	switch column.GetTypeDesc().GetTypes()[0].PrimitiveEntry.GetType() {
	case thriftlib.TTypeId_BOOLEAN_TYPE:
		if !value.IsSetBoolVal() {
			return nil, errors.New("hive: column is bool but value has no bool")
		}

		return value.GetBoolVal().GetValues(), nil
	case thriftlib.TTypeId_TINYINT_TYPE:
		if !value.IsSetByteVal() {
			return nil, errors.New("hive: column is byte but value has no byte")
		}

		return value.GetByteVal().GetValues(), nil
	case thriftlib.TTypeId_SMALLINT_TYPE:
		if !value.IsSetI16Val() {
			return nil, errors.New("hive: column is i16 but value has no i16")
		}

		return value.GetI16Val().GetValues(), nil
	case thriftlib.TTypeId_INT_TYPE:
		if !value.IsSetI32Val() {
			return nil, errors.New("hive: column is i32 but value has no i32")
		}

		return value.GetI32Val().GetValues(), nil
	case thriftlib.TTypeId_BIGINT_TYPE, thriftlib.TTypeId_TIMESTAMP_TYPE:
		if !value.IsSetI64Val() {
			return nil, errors.New("hive: column is i64/timestamp but value has no i16")
		}

		return value.GetI64Val().GetValues(), nil
	case thriftlib.TTypeId_DOUBLE_TYPE, thriftlib.TTypeId_FLOAT_TYPE:
		if !value.IsSetDoubleVal() {
			return nil, errors.New("hive: column is double/float type but value has no double")
		}

		return value.GetDoubleVal().GetValues(), nil
	case thriftlib.TTypeId_STRING_TYPE, thriftlib.TTypeId_MAP_TYPE, thriftlib.TTypeId_STRUCT_TYPE, thriftlib.TTypeId_UNION_TYPE, thriftlib.TTypeId_BINARY_TYPE, thriftlib.TTypeId_DECIMAL_TYPE, thriftlib.TTypeId_NULL_TYPE, thriftlib.TTypeId_INTERVAL_YEAR_MONTH_TYPE, thriftlib.TTypeId_INTERVAL_DAY_TIME_TYPE:
		if !value.IsSetStringVal() {
			return nil, errors.New("hive: column is string but value has no string")
		}

		return value.GetStringVal().GetValues(), nil
	default:
		return nil, errors.New("hive: unknown type from column: " + strconv.Itoa(int(column.GetTypeDesc().GetTypes()[0].PrimitiveEntry.GetType())))
	}
}
