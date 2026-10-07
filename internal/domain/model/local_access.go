package model

import "time"

// Apertura sin conexión (contrato K3 de concepts-apertura-local, vouchers del flujo de pago).

// LockerConnectivity: si el rack está sin conexión y si se le puede pedir un voucher.
type LockerConnectivity struct {
	Online                  bool
	LastSeenAt              *time.Time
	OfflineVoucherAvailable bool
}

// OfflineVoucher: la URL se abre como navegación (nunca con fetch) estando en el Wi-Fi del rack.
type OfflineVoucher struct {
	URL        string
	WifiSSID   string
	ValidUntil time.Time
}
