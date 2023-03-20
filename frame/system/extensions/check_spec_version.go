package system

import (
	sc "github.com/LimeChain/goscale"
	"github.com/LimeChain/gosemble/constants"
	"github.com/LimeChain/gosemble/execution/types"
	primitives "github.com/LimeChain/gosemble/primitives/types"
)

type CheckSpecVersion struct{}

func (_ CheckSpecVersion) AdditionalSigned() (ok sc.U32, err primitives.TransactionValidityError) {
	return constants.RuntimeVersion.SpecVersion, err
}

func (_ CheckSpecVersion) Validate(_who *primitives.Address32, _call *types.Call, _info *primitives.DispatchInfo, _length sc.Compact) (ok primitives.ValidTransaction, err primitives.TransactionValidityError) {
	ok = primitives.DefaultValidTransaction()
	return ok, err
}

func (v CheckSpecVersion) PreDispatch(who *primitives.Address32, call *types.Call, info *primitives.DispatchInfo, length sc.Compact) (ok primitives.Pre, err primitives.TransactionValidityError) {
	_, err = v.Validate(who, call, info, length)
	return ok, err
}
