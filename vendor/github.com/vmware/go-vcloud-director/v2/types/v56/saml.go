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
//
// Note. This structure is not complete and has many more fields.
type VcdSamlMetadata struct {
	XMLName xml.Name `xml:"EntityDescriptor"`
	Text    string   `xml:",chardata"`
	ID      string   `xml:"ID,attr"`
	// EntityID is the configured vCD Entity ID which is used in ADFS authentication request
	EntityID string `xml:"entityID,attr"`
}

// AdfsAuthErrorEnvelope helps to parse ADFS authentication error with help of Error() method
//
// Note. This structure is not complete and has many more fields.
type AdfsAuthErrorEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
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

// AdfsAuthResponseEnvelope helps to marshal ADFS reponse to authentication request.
//
// Note. This structure is not complete and has many more fields.
type AdfsAuthResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		RequestSecurityTokenResponseCollection struct {
			RequestSecurityTokenResponse struct {
				// RequestedSecurityTokenTxt returns data which is accepted by vCD as a SIGN token
				RequestedSecurityTokenTxt InnerXML `xml:"RequestedSecurityToken"`
			} `xml:"RequestSecurityTokenResponse"`
		} `xml:"RequestSecurityTokenResponseCollection"`
	} `xml:"Body"`
}
