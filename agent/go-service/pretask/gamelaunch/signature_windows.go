package gamelaunch

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	crypt32          = windows.NewLazySystemDLL("crypt32.dll")
	cryptMsgGetParam = crypt32.NewProc("CryptMsgGetParam")
	cryptMsgClose    = crypt32.NewProc("CryptMsgClose")
)

// 国服 1.6.3 的 Launcher 规则；证书轮换时需要重新核实，不能降级为仅检查文件名。
var officialFingerprints = map[string]bool{
	"2b366f397e64a9494a553b70004e8ca1aae02b9d": true,
	"ff4af264c36a27f0c805936708ef8ec67b396225": true,
}

// verifyLauncher 返回经 WinVerifyTrust 验证的实际签名者指纹，而非证书库中任意证书。
func verifyLauncher(path string) (string, error) {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	// 核验期间禁止替换或写入文件，保证签名验证与签名者读取针对同一份字节。
	handle, err := windows.CreateFile(path16, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)
	file := windows.WinTrustFileInfo{Size: uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})), FilePath: path16, File: handle}
	// 指定主签名，保证下面读取的 CMSG_SIGNER_INFO 与被验证的签名一致。
	signature := windows.WinTrustSignatureSettings{Size: uint32(unsafe.Sizeof(windows.WinTrustSignatureSettings{})), Flags: 1}
	data := windows.WinTrustData{
		Size:     uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice: windows.WTD_UI_NONE, UnionChoice: windows.WTD_CHOICE_FILE,
		StateAction:                     windows.WTD_STATEACTION_VERIFY,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&file),
		SignatureSettings:               &signature,
	}
	defer func() {
		data.StateAction = windows.WTD_STATEACTION_CLOSE
		_ = windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, &data)
	}()
	if err := windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, &data); err != nil {
		return "", fmt.Errorf("WinVerifyTrust %s: %w", path, err)
	}
	var store, message windows.Handle
	if err := windows.CryptQueryObject(windows.CERT_QUERY_OBJECT_FILE, unsafe.Pointer(path16),
		windows.CERT_QUERY_CONTENT_FLAG_PKCS7_SIGNED_EMBED, windows.CERT_QUERY_FORMAT_FLAG_BINARY,
		0, nil, nil, nil, &store, &message, nil); err != nil {
		return "", fmt.Errorf("read launcher signature: %w", err)
	}
	defer windows.CertCloseStore(store, 0)
	defer cryptMsgClose.Call(uintptr(message))
	// CMSG_SIGNER_INFO_PARAM=6，读取主签名者的 Issuer + SerialNumber，不能遍历任意证书冒充签名者。
	var size uint32
	if ok, _, err := cryptMsgGetParam.Call(uintptr(message), 6, 0, 0, uintptr(unsafe.Pointer(&size))); ok == 0 {
		return "", fmt.Errorf("get signer info size: %w", err)
	}
	type signerInfo struct {
		Version      uint32
		Issuer       windows.CertNameBlob
		SerialNumber windows.CryptIntegerBlob
	}
	if uintptr(size) < unsafe.Sizeof(signerInfo{}) {
		return "", fmt.Errorf("incomplete signer info: %d bytes", size)
	}
	buffer := make([]byte, size)
	if ok, _, err := cryptMsgGetParam.Call(uintptr(message), 6, 0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size))); ok == 0 {
		return "", fmt.Errorf("read signer info: %w", err)
	}
	signer := (*signerInfo)(unsafe.Pointer(&buffer[0]))
	info := windows.CertInfo{Issuer: signer.Issuer, SerialNumber: signer.SerialNumber}
	context, err := windows.CertFindCertificateInStore(store, windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING,
		0, windows.CERT_FIND_SUBJECT_CERT, unsafe.Pointer(&info), nil)
	if err != nil {
		return "", fmt.Errorf("find signer certificate: %w", err)
	}
	defer windows.CertFreeCertificateContext(context)
	if context == nil || context.EncodedCert == nil || context.Length == 0 {
		return "", fmt.Errorf("empty signer certificate for %s", path)
	}
	hash := sha1.Sum(unsafe.Slice(context.EncodedCert, context.Length))
	fingerprint := hex.EncodeToString(hash[:])
	if !officialFingerprints[fingerprint] {
		return "", fmt.Errorf("unrecognized launcher signer %s: %s", fingerprint, path)
	}
	return fingerprint, nil
}
