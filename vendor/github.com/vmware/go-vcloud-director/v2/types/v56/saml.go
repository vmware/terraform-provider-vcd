/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import (
	"encoding/xml"
	"fmt"
)

// VcdSamlMetadata helps to marshal vCD SAML Metadata endpoint response
// https://1.1.1.1/cloud/org/my-org/saml/metadata/alias/vcd
// This body is huge, although only documented fields are used.
type VcdSamlMetadata struct {
	XMLName xml.Name `xml:"EntityDescriptor"`
	Text    string   `xml:",chardata"`
	ID      string   `xml:"ID,attr"`
	// EntityID is the configured vCD Entity ID which is used in ADFS authentication request
	EntityID        string `xml:"entityID,attr"`
	Md              string `xml:"md,attr"`
	SPSSODescriptor struct {
		Text                       string `xml:",chardata"`
		AuthnRequestsSigned        string `xml:"AuthnRequestsSigned,attr"`
		WantAssertionsSigned       string `xml:"WantAssertionsSigned,attr"`
		ProtocolSupportEnumeration string `xml:"protocolSupportEnumeration,attr"`
		KeyDescriptor              []struct {
			Text    string `xml:",chardata"`
			Use     string `xml:"use,attr"`
			KeyInfo struct {
				Text     string `xml:",chardata"`
				Ds       string `xml:"ds,attr"`
				X509Data struct {
					Text            string `xml:",chardata"`
					X509Certificate string `xml:"X509Certificate"`
				} `xml:"X509Data"`
			} `xml:"KeyInfo"`
		} `xml:"KeyDescriptor"`
		SingleLogoutService []struct {
			Text     string `xml:",chardata"`
			Binding  string `xml:"Binding,attr"`
			Location string `xml:"Location,attr"`
		} `xml:"SingleLogoutService"`
		NameIDFormat             []string `xml:"NameIDFormat"`
		AssertionConsumerService []struct {
			Text            string `xml:",chardata"`
			Binding         string `xml:"Binding,attr"`
			Location        string `xml:"Location,attr"`
			Index           string `xml:"index,attr"`
			IsDefault       string `xml:"isDefault,attr"`
			ProtocolBinding string `xml:"ProtocolBinding,attr"`
			Hoksso          string `xml:"hoksso,attr"`
		} `xml:"AssertionConsumerService"`
	} `xml:"SPSSODescriptor"`
}

// AdfsAuthErrorEnvelope helps to parse ADFS authentication error with help of Error() method
type AdfsAuthErrorEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"s,attr"`
	A       string   `xml:"a,attr"`
	U       string   `xml:"u,attr"`
	Header  struct {
		Text   string `xml:",chardata"`
		Action struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
		} `xml:"Action"`
		Security struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
			O              string `xml:"o,attr"`
			Timestamp      struct {
				Text    string `xml:",chardata"`
				ID      string `xml:"Id,attr"`
				Created string `xml:"Created"`
				Expires string `xml:"Expires"`
			} `xml:"Timestamp"`
		} `xml:"Security"`
	} `xml:"Header"`
	Body struct {
		Text  string `xml:",chardata"`
		Fault struct {
			Text string `xml:",chardata"`
			Code struct {
				Text    string `xml:",chardata"`
				Value   string `xml:"Value"`
				Subcode struct {
					Text  string `xml:",chardata"`
					Value struct {
						Text string `xml:",chardata"`
						A    string `xml:"a,attr"`
					} `xml:"Value"`
				} `xml:"Subcode"`
			} `xml:"Code"`
			Reason struct {
				Chardata string `xml:",chardata"`
				Text     struct {
					Text string `xml:",chardata"`
					Lang string `xml:"lang,attr"`
				} `xml:"Text"`
			} `xml:"Reason"`
		} `xml:"Fault"`
	} `xml:"Body"`
}

// Error satisfies Go's default `error` interface for AdfsAuthErrorEnvelope and formats
// error for humand readable output
func (samlErr AdfsAuthErrorEnvelope) Error() string {
	return fmt.Sprintf("SAML request got error: %s", samlErr.Body.Fault.Reason.Text)
}

// AdfsAuthResponseEnvelope helps to marshal ADFS reponse to authentication request
type AdfsAuthResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"s,attr"`
	A       string   `xml:"a,attr"`
	U       string   `xml:"u,attr"`
	Header  struct {
		Text   string `xml:",chardata"`
		Action struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
		} `xml:"Action"`
		Security struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
			O              string `xml:"o,attr"`
			Timestamp      struct {
				Text    string `xml:",chardata"`
				ID      string `xml:"Id,attr"`
				Created string `xml:"Created"`
				Expires string `xml:"Expires"`
			} `xml:"Timestamp"`
		} `xml:"Security"`
	} `xml:"Header"`
	Body struct {
		Text                                   string `xml:",chardata"`
		RequestSecurityTokenResponseCollection struct {
			Text                         string `xml:",chardata"`
			Trust                        string `xml:"trust,attr"`
			RequestSecurityTokenResponse struct {
				Text     string `xml:",chardata"`
				Lifetime struct {
					Text    string `xml:",chardata"`
					Created struct {
						Text string `xml:",chardata"`
						Wsu  string `xml:"wsu,attr"`
					} `xml:"Created"`
					Expires struct {
						Text string `xml:",chardata"`
						Wsu  string `xml:"wsu,attr"`
					} `xml:"Expires"`
				} `xml:"Lifetime"`
				AppliesTo struct {
					Text              string `xml:",chardata"`
					Wsp               string `xml:"wsp,attr"`
					EndpointReference struct {
						Text    string `xml:",chardata"`
						Wsa     string `xml:"wsa,attr"`
						Address string `xml:"Address"`
					} `xml:"EndpointReference"`
				} `xml:"AppliesTo"`
				// RequestedSecurityTokenTxt returns data which is accepted by vCD as a SIGN token
				RequestedSecurityTokenTxt InnerXML `xml:"RequestedSecurityToken"`
				RequestedDisplayToken     struct {
					Text         string `xml:",chardata"`
					I            string `xml:"i,attr"`
					DisplayToken struct {
						Text         string `xml:",chardata"`
						Lang         string `xml:"lang,attr"`
						DisplayClaim []struct {
							Text         string `xml:",chardata"`
							URI          string `xml:"Uri,attr"`
							DisplayTag   string `xml:"DisplayTag"`
							Description  string `xml:"Description"`
							DisplayValue string `xml:"DisplayValue"`
						} `xml:"DisplayClaim"`
					} `xml:"DisplayToken"`
				} `xml:"RequestedDisplayToken"`
				RequestedAttachedReference struct {
					Text                   string `xml:",chardata"`
					SecurityTokenReference struct {
						Text          string `xml:",chardata"`
						TokenType     string `xml:"TokenType,attr"`
						Xmlns         string `xml:"xmlns,attr"`
						B             string `xml:"b,attr"`
						KeyIdentifier struct {
							Text      string `xml:",chardata"`
							ValueType string `xml:"ValueType,attr"`
						} `xml:"KeyIdentifier"`
					} `xml:"SecurityTokenReference"`
				} `xml:"RequestedAttachedReference"`
				RequestedUnattachedReference struct {
					Text                   string `xml:",chardata"`
					SecurityTokenReference struct {
						Text          string `xml:",chardata"`
						TokenType     string `xml:"TokenType,attr"`
						Xmlns         string `xml:"xmlns,attr"`
						B             string `xml:"b,attr"`
						KeyIdentifier struct {
							Text      string `xml:",chardata"`
							ValueType string `xml:"ValueType,attr"`
						} `xml:"KeyIdentifier"`
					} `xml:"SecurityTokenReference"`
				} `xml:"RequestedUnattachedReference"`
				TokenType   string `xml:"TokenType"`
				RequestType string `xml:"RequestType"`
				KeyType     string `xml:"KeyType"`
			} `xml:"RequestSecurityTokenResponse"`
		} `xml:"RequestSecurityTokenResponseCollection"`
	} `xml:"Body"`
}
