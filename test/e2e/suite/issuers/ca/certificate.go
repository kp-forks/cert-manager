/*
Copyright 2020 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ca

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cert-manager/cert-manager/e2e-tests/framework"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/helper/validation/certificates"
	"github.com/cert-manager/cert-manager/e2e-tests/util"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/test/unit/gen"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = framework.CertManagerDescribe("CA Certificate", func() {
	f := framework.NewDefaultFramework("create-ca-certificate")

	issuerName := "test-ca-issuer"
	issuerSecretName := "ca-issuer-signing-keypair"
	certificateName := "test-ca-certificate"
	certificateSecretName := "test-ca-certificate"

	JustBeforeEach(func(testingCtx context.Context) {
		By("Creating an Issuer")
		issuer := gen.Issuer(issuerName,
			gen.SetIssuerNamespace(f.Namespace.Name),
			gen.SetIssuerCASecretName(issuerSecretName))
		_, err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Create(testingCtx, issuer, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		By("Waiting for Issuer to become Ready")
		err = util.WaitForIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name),
			issuerName,
			v1.IssuerCondition{
				Type:   v1.IssuerConditionReady,
				Status: cmmeta.ConditionTrue,
			})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func(testingCtx context.Context) {
		By("Cleaning up")
		err := f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Delete(testingCtx, issuerSecretName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
		err = f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Delete(testingCtx, issuerName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the CA is the root", func() {
		BeforeEach(func(testingCtx context.Context) {
			By("Creating a signing keypair fixture")
			_, err := f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Create(testingCtx, newSigningKeypairSecret(issuerSecretName), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should generate a signed keypair", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate")
			cert := gen.Certificate(certificateName,
				gen.SetCertificateNamespace(f.Namespace.Name),
				gen.SetCertificateSecretName(certificateSecretName),
				gen.SetCertificateIssuer(cmmeta.ObjectReference{
					Name: issuerName,
					Kind: v1.IssuerKind,
				}),
				gen.SetCertificateCommonName("test.domain.com"),
				gen.SetCertificateOrganization("test-org"),
			)
			cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Verifying the Certificate is valid")
			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be able to obtain an ECDSA key from a RSA backed issuer", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate")
			cert := gen.Certificate(certificateName,
				gen.SetCertificateNamespace(f.Namespace.Name),
				gen.SetCertificateSecretName(certificateSecretName),
				gen.SetCertificateIssuer(cmmeta.ObjectReference{
					Name: issuerName,
					Kind: v1.IssuerKind,
				}),
				gen.SetCertificateCommonName("test.domain.com"),
				gen.SetCertificateOrganization("test-org"),
				gen.SetCertificateKeyAlgorithm(v1.ECDSAKeyAlgorithm),
				gen.SetCertificateKeySize(521),
			)
			cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be able to obtain an Ed25519 key from a RSA backed issuer", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate")
			cert := gen.Certificate(certificateName,
				gen.SetCertificateNamespace(f.Namespace.Name),
				gen.SetCertificateSecretName(certificateSecretName),
				gen.SetCertificateIssuer(cmmeta.ObjectReference{
					Name: issuerName,
					Kind: v1.IssuerKind,
				}),
				gen.SetCertificateCommonName("test.domain.com"),
				gen.SetCertificateOrganization("test-org"),
				gen.SetCertificateKeyAlgorithm(v1.Ed25519KeyAlgorithm),
			)
			cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be able to create a certificate with additional output formats", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate")
			cert := gen.Certificate(certificateName,
				gen.SetCertificateNamespace(f.Namespace.Name),
				gen.SetCertificateSecretName(certificateSecretName),
				gen.SetCertificateIssuer(cmmeta.ObjectReference{
					Name: issuerName,
					Kind: v1.IssuerKind,
				}),
				gen.SetCertificateCommonName("test.domain.com"),
				gen.SetCertificateOrganization("test-org"),
				gen.SetCertificateAdditionalOutputFormats(
					v1.CertificateAdditionalOutputFormat{Type: "DER"},
					v1.CertificateAdditionalOutputFormat{Type: "CombinedPEM"},
				),
			)
			cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})

		cases := []struct {
			inputDuration    *metav1.Duration
			inputRenewBefore *metav1.Duration
			expectedDuration time.Duration
			label            string
		}{
			{
				inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 35},
				inputRenewBefore: nil,
				expectedDuration: time.Hour * 24 * 35,
				label:            "35 days",
			},
			{
				inputDuration:    nil,
				inputRenewBefore: nil,
				expectedDuration: time.Hour * 24 * 90,
				label:            "the default duration (90 days)",
			},
		}
		for _, v := range cases {
			It("should generate a signed keypair valid for "+v.label, func(testingCtx context.Context) {
				certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

				By("Creating a Certificate")
				cert := gen.Certificate(certificateName,
					gen.SetCertificateNamespace(f.Namespace.Name),
					gen.SetCertificateSecretName(certificateSecretName),
					gen.SetCertificateIssuer(cmmeta.ObjectReference{
						Name: issuerName,
						Kind: v1.IssuerKind,
					}),
					gen.SetCertificateDuration(v.inputDuration),
					gen.SetCertificateRenewBefore(v.inputRenewBefore),
					gen.SetCertificateCommonName("test.domain.com"),
					gen.SetCertificateOrganization("test-org"),
				)
				cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting for the Certificate to be issued...")
				cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
				Expect(err).NotTo(HaveOccurred())

				By("Validating the issued Certificate...")
				err = f.Helper().ValidateCertificate(cert)
				Expect(err).NotTo(HaveOccurred())

				err = f.Helper().ValidateCertificate(cert, certificates.ExpectDuration(v.expectedDuration, 0))
				Expect(err).NotTo(HaveOccurred())
			})
		}
	})

	Context("when the CA is an issuer", func() {
		BeforeEach(func(testingCtx context.Context) {
			By("Creating a signing keypair fixture")
			_, err := f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Create(testingCtx, newSigningIssuer1KeypairSecret(issuerSecretName), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should generate a signed keypair", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate")
			cert := gen.Certificate(certificateName,
				gen.SetCertificateNamespace(f.Namespace.Name),
				gen.SetCertificateSecretName(certificateSecretName),
				gen.SetCertificateIssuer(cmmeta.ObjectReference{
					Name: issuerName,
					Kind: v1.IssuerKind,
				}),
				gen.SetCertificateCommonName("test.domain.com"),
				gen.SetCertificateOrganization("test-org"),
			)
			cert, err := certClient.Create(testingCtx, cert, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the CA is a second level issuer", func() {
		BeforeEach(func(testingCtx context.Context) {
			By("Creating a signing keypair fixture")
			_, err := f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Create(testingCtx, newSigningIssuer2KeypairSecret(issuerSecretName), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should generate a signed keypair", func(testingCtx context.Context) {
			certClient := f.CertManagerClientSet.CertmanagerV1().Certificates(f.Namespace.Name)

			By("Creating a Certificate with Usages")
			cert, err := certClient.Create(testingCtx, gen.Certificate(certificateName, gen.SetCertificateNamespace(f.Namespace.Name), gen.SetCertificateCommonName("test.domain.com"), gen.SetCertificateSecretName(certificateSecretName), gen.SetCertificateIssuer(cmmeta.ObjectReference{Name: issuerName, Kind: v1.IssuerKind}), gen.SetCertificateKeyUsages(v1.UsageServerAuth, v1.UsageClientAuth)), metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Waiting for the Certificate to be issued...")
			cert, err = f.Helper().WaitForCertificateReadyAndDoneIssuing(testingCtx, cert, time.Minute*5)
			Expect(err).NotTo(HaveOccurred())

			By("Validating the issued Certificate...")
			err = f.Helper().ValidateCertificate(cert)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
