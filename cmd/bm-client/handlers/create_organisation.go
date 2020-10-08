package handlers

import (
	"fmt"
	"os"

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/internal/config"
	"github.com/bitmaelum/bitmaelum-suite/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/internal/organisation"
	"github.com/bitmaelum/bitmaelum-suite/internal/vault"
	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	pow "github.com/bitmaelum/bitmaelum-suite/pkg/proofofwork"
)

// CreateOrganisation creates a new organisation locally in the vault and pushes the public key to the resolver
func CreateOrganisation(vault *vault.Vault, orgAddr, fullName string, orgValidations []string) {
	fmt.Printf("* Verifying if organisation name is valid: ")
	orgHash := hash.New(orgAddr)

	fmt.Printf("* Checking if your validations are correct: ")
	val, err := organisation.NewValidationTypeFromStringArray(orgValidations)
	if err != nil {
		fmt.Print("\n  X it seems that one of your validations is wrong: ", err)
		fmt.Println("")
		os.Exit(1)
	}
	fmt.Printf("ok.\n")

	fmt.Printf("* Checking if organisation is already known in the resolver service: ")
	ks := container.GetResolveService()
	_, err = ks.ResolveOrganisation(orgHash)
	if err == nil {
		fmt.Printf("\n  X it seems that this organisation is already in use. Please specify another organisation.")
		fmt.Println("")
		os.Exit(1)
	}
	fmt.Printf("not found. This is a good thing.\n")

	var seed string

	fmt.Printf("* Checking if the organisation is already present in the vault: ")
	var info *internal.OrganisationInfo
	if vault.HasOrganisation(orgHash) {
		fmt.Printf("\n  X organisation already present in the vault.\n")
		fmt.Println("")
		os.Exit(1)
	} else {
		fmt.Printf("not found. This is a good thing.\n")

		fmt.Printf("* Generating organisation public/private key pair: ")

		var (
			privKey *bmcrypto.PrivKey
			pubKey  *bmcrypto.PubKey
		)
		seed, privKey, pubKey, err = bmcrypto.GenerateKeypairWithSeed()
		if err != nil {
			fmt.Print(err)
			fmt.Println("")
			os.Exit(1)
		}
		fmt.Printf("done.\n")

		fmt.Printf("* Doing some work to let people know this is not a fake account: ")
		proof := pow.NewWithoutProof(config.Client.Accounts.ProofOfWork, orgHash.String())
		proof.WorkMulticore()
		fmt.Printf("done.\n")

		fmt.Printf("* Adding your new organisation into the vault: ")
		info = &internal.OrganisationInfo{
			Addr:        orgAddr,
			FullName:    fullName,
			PrivKey:     *privKey,
			PubKey:      *pubKey,
			Pow:         *proof,
			Validations: val,
		}

		vault.AddOrganisation(*info)
		err = vault.WriteToDisk()
		if err != nil {
			fmt.Printf("\n  X error while saving organisation into vault: %#v", err)
			fmt.Println("")
			os.Exit(1)
		}
		fmt.Printf("done\n")
	}

	fmt.Printf("* Making your organisation known to the outside world: ")
	err = ks.UploadOrganisationInfo(*info)
	if err != nil {
		// We can't remove the account from the vault as we have created it on the mail-server

		fmt.Printf("\n  X error while uploading organisation to the resolver: " + err.Error())
		fmt.Printf("\n  X Please try again with:\n   bm-client push-organisation -a '%s'\n", orgHash.String())
		fmt.Println("")
		os.Exit(1)
	}
	fmt.Printf("done\n")

	fmt.Printf("\n")
	fmt.Printf("* All done")

	if len(seed) > 0 {
		fmt.Print(`
*****************************************************************************
!IMPORTANT!IMPORTANT!IMPORTANT!IMPORTANT!IMPORTANT!IMPORTANT!IMPORTANT!IMPORT
*****************************************************************************

We have generated a private key which allows you to control the organisation. 
If, for any reason, you lose this key, you will need to use the following 
words in order to recreate the key:

`)
		fmt.Print(internal.WordWrap(seed, 78))
		fmt.Print(`

Write these words down and store them in a secure environment. They are the 
ONLY way to recover your private key in case you lose it.

WITHOUT THESE WORDS, ALL ACCESS TO YOUR ORGANISATION IS LOST!
`)
	}
}
