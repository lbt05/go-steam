package steamid

// Bits for each portion of SteamID.
const (
	accountIDOffset, accountIDMask             = 00, 0xFFFFFFFF // 32 bits
	accountInstanceOffset, accountInstanceMask = 32, 0xFFFFF    // 20 bits
	accountTypeOffset, accountTypeMask         = 52, 0xF        // 4 bits
	accountUniverseOffset, accountUniverseMask = 56, 0xF        // 4 bits
)

// AccountID returns the 32-bit account ID.
func (sid SteamID) AccountID() AccountID {
	return AccountID(sid.getBits(accountIDMask, accountIDOffset))
}

// WithAccountID updates the 32-bit account ID.
func (sid SteamID) WithAccountID(accountID AccountID) SteamID {
	return sid.setBits(uint64(accountID), accountIDMask, accountIDOffset)
}

// AccountInstance returns the 20-bit account instance.
func (sid SteamID) AccountInstance() AccountInstance {
	return AccountInstance(sid.getBits(accountInstanceMask, accountInstanceOffset))
}

// WithAccountInstance updates the 20-bit account instance.
func (sid SteamID) WithAccountInstance(accountInstance AccountInstance) SteamID {
	return sid.setBits(uint64(accountInstance), accountInstanceMask, accountInstanceOffset)
}

// AccountType returns the 4-bit account type.
func (sid SteamID) AccountType() AccountType {
	return AccountType(sid.getBits(accountTypeMask, accountTypeOffset))
}

// WithAccountType updates the 4-bit account type.
func (sid SteamID) WithAccountType(accountType AccountType) SteamID {
	return sid.setBits(uint64(accountType), accountTypeMask, accountTypeOffset)
}

// AccountUniverse returns the 4-bit account universe.
func (sid SteamID) AccountUniverse() AccountUniverse {
	return AccountUniverse(sid.getBits(accountUniverseMask, accountUniverseOffset))
}

// WithAccountUniverse updates the 4-bit account universe.
func (sid SteamID) WithAccountUniverse(accountUniverse AccountUniverse) SteamID {
	return sid.setBits(uint64(accountUniverse), accountUniverseMask, accountUniverseOffset)
}

// getBits returns the bits at the given offset.
func (sid SteamID) getBits(mask uint64, offset uint64) uint64 {
	return (sid.Uint64() >> offset) & mask
}

// setBits updates the bits at the given offset.
func (sid SteamID) setBits(value uint64, mask uint64, offset uint64) SteamID {
	return SteamID((sid.Uint64() & ^(mask << offset)) | (value&mask)<<offset)
}
