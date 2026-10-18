package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Credential struct {
	CredentialID     string `json:"credentialId"`
	IssuerDID        string `json:"issuerDID"`
	SubjectDID       string `json:"subjectDID"`
	CredentialHash   string `json:"credentialHash"`
	DegreeType       string `json:"degreeType"`
	Major            string `json:"major"`
	University       string `json:"university"`
	GraduationDate   string `json:"graduationDate"`
	Status           string `json:"status"`
	IssuedAt         string `json:"issuedAt"`
	RevocationReason string `json:"revocationReason"`
	RevokedAt        string `json:"revokedAt"`
}

type IssuerRecord struct {
	IssuerDID    string `json:"issuerDID"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	Trusted      bool   `json:"trusted"`
	RegisteredAt string `json:"registeredAt"`
}

type CredentialContract struct {
	contractapi.Contract
}

const credPrefix = "CRED_"
const issuerPrefix = "ISSUER_"

func (c *CredentialContract) RegisterIssuer(ctx contractapi.TransactionContextInterface, did, name, country string) error {
	rec := IssuerRecord{did, name, country, true, time.Now().UTC().Format(time.RFC3339)}
	data, _ := json.Marshal(rec)
	return ctx.GetStub().PutState(issuerPrefix+did, data)
}

func (c *CredentialContract) issuerTrusted(ctx contractapi.TransactionContextInterface, did string) (bool, error) {
	d, err := ctx.GetStub().GetState(issuerPrefix + did)
	if err != nil {
		return false, err
	}
	if d == nil {
		return false, nil
	}
	var r IssuerRecord
	if err := json.Unmarshal(d, &r); err != nil {
		return false, err
	}
	return r.Trusted, nil
}

func (c *CredentialContract) IssueCredential(ctx contractapi.TransactionContextInterface, id, issuerDID, subjectDID, degreeType, major, university, graduationDate, doc string) error {
	trusted, err := c.issuerTrusted(ctx, issuerDID)
	if err != nil {
		return err
	}
	if !trusted {
		return fmt.Errorf("issuer %s not trusted", issuerDID)
	}
	ex, err := ctx.GetStub().GetState(credPrefix + id)
	if err != nil {
		return err
	}
	if ex != nil {
		return fmt.Errorf("credential %s exists", id)
	}
	sum := sha256.Sum256([]byte(doc))
	cred := Credential{id, issuerDID, subjectDID, hex.EncodeToString(sum[:]), degreeType, major, university, graduationDate, "valid", time.Now().UTC().Format(time.RFC3339), "", ""}
	data, _ := json.Marshal(cred)
	return ctx.GetStub().PutState(credPrefix+id, data)
}

func (c *CredentialContract) VerifyCredential(ctx contractapi.TransactionContextInterface, id string) (*Credential, error) {
	d, err := ctx.GetStub().GetState(credPrefix + id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("credential %s not found", id)
	}
	var cred Credential
	if err := json.Unmarshal(d, &cred); err != nil {
		return nil, err
	}
	trusted, err := c.issuerTrusted(ctx, cred.IssuerDID)
	if err != nil {
		return nil, err
	}
	if !trusted {
		cred.Status = "issuer_untrusted"
	}
	return &cred, nil
}

func (c *CredentialContract) RevokeCredential(ctx contractapi.TransactionContextInterface, id, reason string) error {
	ok := map[string]bool{"academic_misconduct": true, "administrative_error": true, "degree_withdrawn": true}
	if !ok[reason] {
		return fmt.Errorf("invalid reason: %s", reason)
	}
	d, err := ctx.GetStub().GetState(credPrefix + id)
	if err != nil {
		return err
	}
	if d == nil {
		return fmt.Errorf("credential %s not found", id)
	}
	var cred Credential
	if err := json.Unmarshal(d, &cred); err != nil {
		return err
	}
	if cred.Status == "revoked" {
		return fmt.Errorf("already revoked")
	}
	cred.Status = "revoked"
	cred.RevocationReason = reason
	cred.RevokedAt = time.Now().UTC().Format(time.RFC3339)
	data, _ := json.Marshal(cred)
	return ctx.GetStub().PutState(credPrefix+id, data)
}

func (c *CredentialContract) AssessSubstantialDifferences(ctx contractapi.TransactionContextInterface, degreeType string, years int) (string, error) {
	if (degreeType == "associate" || degreeType == "bachelor") && years < 4 {
		return "manual_review", nil
	}
	return "equivalent", nil
}

func main() {
	cc, err := contractapi.NewChaincode(&CredentialContract{})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting: %s\n", err)
	}
}
