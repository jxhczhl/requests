package transport

import (
	"crypto/sha256"
	"strconv"
	"strings"

	utls "github.com/refraction-networking/utls"
)

const (
	chrome  = "chrome"  //chrome User agent enum
	firefox = "firefox" //firefox User agent enum
)

func parseUserAgent(userAgent string) string {
	switch {
	case strings.Contains(strings.ToLower(userAgent), "chrome"):
		return chrome
	case strings.Contains(strings.ToLower(userAgent), "firefox"):
		return firefox
	default:
		return chrome
	}

}

// StringToSpec creates a ClientHelloSpec based on a JA3 string
func StringToSpec(ja3 string, userAgent string, tlsExtensions *TLSExtensions, forceHTTP1 bool) (*utls.ClientHelloSpec, error) {
	parsedUserAgent := parseUserAgent(userAgent)
	ext := tlsExtensions
	extMap := genMap()
	tokens := strings.Split(ja3, ",")

	version := tokens[0]
	ciphers := strings.Split(tokens[1], "-")
	extensions := strings.Split(tokens[2], "-")
	curves := strings.Split(tokens[3], "-")
	if len(curves) == 1 && curves[0] == "" {
		curves = []string{}
	}
	pointFormats := strings.Split(tokens[4], "-")
	if len(pointFormats) == 1 && pointFormats[0] == "" {
		pointFormats = []string{}
	}
	// parse curves
	var targetCurves []utls.CurveID
	if parsedUserAgent == chrome && !tlsExtensions.NotUsedGREASE {
		targetCurves = append(targetCurves, utls.CurveID(utls.GREASE_PLACEHOLDER)) //append grease for Chrome browsers
	}
	for _, c := range curves {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		targetCurves = append(targetCurves, utls.CurveID(cid))
	}
	extMap["10"] = &utls.SupportedCurvesExtension{Curves: targetCurves}

	// parse point formats
	var targetPointFormats []byte
	for _, p := range pointFormats {
		pid, err := strconv.ParseUint(p, 10, 8)
		if err != nil {
			return nil, err
		}
		targetPointFormats = append(targetPointFormats, byte(pid))
	}
	extMap["11"] = &utls.SupportedPointsExtension{SupportedPoints: targetPointFormats}

	// force http1
	if forceHTTP1 {
		extMap["16"] = &utls.ALPNExtension{
			AlpnProtocols: []string{"http/1.1"},
		}
	}

	// custom tls extensions
	if tlsExtensions != nil {
		if ext.SupportedSignatureAlgorithms != nil {
			extMap["13"] = ext.SupportedSignatureAlgorithms
		}
		if ext.CertCompressionAlgo != nil {
			extMap["27"] = ext.CertCompressionAlgo
		}
		if ext.RecordSizeLimit != nil {
			extMap["28"] = ext.RecordSizeLimit
		}
		if ext.DelegatedCredentials != nil {
			extMap["34"] = ext.DelegatedCredentials
		}
		if ext.SupportedVersions != nil {
			extMap["43"] = ext.SupportedVersions
		}
		if ext.PSKKeyExchangeModes != nil {
			extMap["45"] = ext.PSKKeyExchangeModes
		}
		if ext.SignatureAlgorithmsCert != nil {
			extMap["50"] = ext.SignatureAlgorithmsCert
		}
		if ext.KeyShareCurves != nil {
			extMap["51"] = ext.KeyShareCurves
		}
	}

	// set extension 43
	vid64, err := strconv.ParseUint(version, 10, 16)
	if err != nil {
		return nil, err
	}
	vid := uint16(vid64)

	// build extenions list
	var exts []utls.TLSExtension
	//Optionally Add Chrome Grease Extension
	if parsedUserAgent == chrome && !tlsExtensions.NotUsedGREASE {
		exts = append(exts, &utls.UtlsGREASEExtension{})
	}
	for _, e := range extensions {
		te, ok := extMap[e]
		if !ok {
			return nil, raiseExtensionError(e)
		}
		// //Optionally add Chrome Grease Extension
		if e == "21" && parsedUserAgent == chrome && !tlsExtensions.NotUsedGREASE {
			exts = append(exts, &utls.UtlsGREASEExtension{})
		}
		exts = append(exts, te)
	}

	// build CipherSuites
	var suites []uint16
	//Optionally Add Chrome Grease Extension
	if parsedUserAgent == chrome && !tlsExtensions.NotUsedGREASE {
		suites = append(suites, utls.GREASE_PLACEHOLDER)
	}
	for _, c := range ciphers {
		cid, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return nil, err
		}
		suites = append(suites, uint16(cid))
	}
	_ = vid
	return &utls.ClientHelloSpec{
		// TLSVersMin:         vid,
		// TLSVersMax:         vid,
		CipherSuites:       suites,
		CompressionMethods: []byte{0},
		Extensions:         exts,
		GetSessionID:       sha256.Sum256,
	}, nil
}

func genMap() (extMap map[string]utls.TLSExtension) {
	extMap = map[string]utls.TLSExtension{
		"0": &utls.SNIExtension{},
		"5": &utls.StatusRequestExtension{},
		// These are applied later
		// "10": &tls.SupportedCurvesExtension{...}
		// "11": &tls.SupportedPointsExtension{...}
		"13": &utls.SignatureAlgorithmsExtension{
			SupportedSignatureAlgorithms: []utls.SignatureScheme{
				utls.ECDSAWithP256AndSHA256,
				utls.ECDSAWithP384AndSHA384,
				utls.ECDSAWithP521AndSHA512,
				utls.PSSWithSHA256,
				utls.PSSWithSHA384,
				utls.PSSWithSHA512,
				utls.PKCS1WithSHA256,
				utls.PKCS1WithSHA384,
				utls.PKCS1WithSHA512,
				utls.ECDSAWithSHA1,
				utls.PKCS1WithSHA1,
			},
		},
		"16": &utls.ALPNExtension{
			AlpnProtocols: []string{"h2", "http/1.1"},
		},
		"17": &utls.GenericExtension{Id: 17}, // status_request_v2
		"18": &utls.SCTExtension{},
		"21": &utls.UtlsPaddingExtension{GetPaddingLen: utls.BoringPaddingStyle},
		"22": &utls.GenericExtension{Id: 22}, // encrypt_then_mac
		"23": &utls.UtlsExtendedMasterSecretExtension{},
		"27": &utls.UtlsCompressCertExtension{
			Algorithms: []utls.CertCompressionAlgo{utls.CertCompressionBrotli},
		},
		"28": &utls.FakeRecordSizeLimitExtension{
			Limit: 0x4001,
		}, //Limit: 0x4001
		"34": &utls.DelegatedCredentialsExtension{
			SupportedSignatureAlgorithms: []utls.SignatureScheme{
				utls.ECDSAWithP256AndSHA256,
				utls.ECDSAWithP384AndSHA384,
				utls.ECDSAWithP521AndSHA512,
				utls.ECDSAWithSHA1,
			},
		},
		"35": &utls.SessionTicketExtension{},
		//"41": &utls.GenericExtension{Id: 41}, //FIXME pre_shared_key, Currently not supported 41 extension
		"43": &utls.SupportedVersionsExtension{Versions: []uint16{
			utls.GREASE_PLACEHOLDER,
			utls.VersionTLS13,
			utls.VersionTLS12,
			utls.VersionTLS11,
			utls.VersionTLS10,
		}},
		"44": &utls.CookieExtension{},
		"45": &utls.PSKKeyExchangeModesExtension{Modes: []uint8{
			utls.PskModeDHE,
		}},
		"49": &utls.GenericExtension{Id: 49}, // post_handshake_auth
		"50": &utls.SignatureAlgorithmsCertExtension{
			SupportedSignatureAlgorithms: []utls.SignatureScheme{
				utls.ECDSAWithP256AndSHA256,
				utls.ECDSAWithP384AndSHA384,
				utls.ECDSAWithP521AndSHA512,
				utls.PSSWithSHA256,
				utls.PSSWithSHA384,
				utls.PSSWithSHA512,
				utls.PKCS1WithSHA256,
				utls.PKCS1WithSHA384,
				utls.PKCS1WithSHA512,
				utls.ECDSAWithSHA1,
				utls.PKCS1WithSHA1,
			},
		}, // signature_algorithms_cert
		"51": &utls.KeyShareExtension{KeyShares: []utls.KeyShare{
			{Group: utls.CurveID(utls.GREASE_PLACEHOLDER), Data: []byte{0}},
			{Group: utls.X25519},

			// {Group: utls.CurveP384}, known bug missing correct extensions for handshake
		}},
		"13172": &utls.NPNExtension{},
		"17513": &utls.ApplicationSettingsExtension{
			SupportedProtocols: []string{
				"h2",
			},
		},
		"30032": &utls.GenericExtension{Id: 0x7550, Data: []byte{0}}, //FIXME
		"65281": &utls.RenegotiationInfoExtension{
			Renegotiation: utls.RenegotiateOnceAsClient,
		},
	}
	return

}
