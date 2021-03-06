// Copyright © 2019 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pairing

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/url"
	"path"
	"time"
)

// pairingCmd represents the pairing command
var PairingCmd = &cobra.Command{
	Use:               "pairing",
	Short:             "Interact with Pairing API",
	Long:              `Interact with pairing API to register devices or to work with device credentials`,
	PersistentPreRunE: pairingPersistentPreRunE,
}

var realm string
var pairingJwt string
var pairingUrl string

func init() {
	PairingCmd.PersistentFlags().StringP("realm-key", "k", "",
		"Path to realm private key used to generate JWT for authentication")
	PairingCmd.MarkPersistentFlagFilename("realm-key")
	viper.BindPFlag("realm.key", PairingCmd.PersistentFlags().Lookup("realm-key"))
	PairingCmd.PersistentFlags().String("pairing-url", "",
		"Pairing API base URL. Defaults to <astarte-url>/pairing.")
	viper.BindPFlag("pairing.url", PairingCmd.PersistentFlags().Lookup("pairing-url"))
	PairingCmd.PersistentFlags().StringP("realm-name", "r", "",
		"The name of the realm that will be queried")
	viper.BindPFlag("realm.name", PairingCmd.PersistentFlags().Lookup("realm-name"))
}

func pairingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	pairingUrlOverride := viper.GetString("pairing.url")
	astarteUrl := viper.GetString("url")
	if pairingUrlOverride != "" {
		// Use explicit pairing-url
		pairingUrl = pairingUrlOverride
	} else if astarteUrl != "" {
		url, _ := url.Parse(astarteUrl)
		url.Path = path.Join(url.Path, "pairing")
		pairingUrl = url.String()
	} else {
		return errors.New("Either astarte-url or pairing-url have to be specified")
	}

	pairingKey := viper.GetString("realm.key")
	if pairingKey == "" {
		return errors.New("realm-key is required")
	}

	realm = viper.GetString("realm.name")
	if realm == "" {
		return errors.New("realm is required")
	}

	var err error
	pairingJwt, err = generatePairingJWT(pairingKey)
	if err != nil {
		return err
	}

	return nil
}

func generatePairingJWT(privateKey string) (jwtString string, err error) {
	keyPEM, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyPEM)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Unix()
	// 5 minutes expiry
	expiry := now + 300
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"a_pa": []string{"^.*$::^.*$"},
		"iat":  now,
		"exp":  expiry,
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
