#!/bin/bash

# create a self-signed certificate
openssl req -x509 -newkey rsa:2048 -keyout tls.key -out tls.crt -days 365 \
  -nodes -subj "/CN=manba-validation-webhook.manba.svc"
# create a secret out of this self-signed cert-key pair
kubectl create secret tls manba-validation-webhook -n manba \
      --key tls.key --cert tls.crt
# enable the Admission Webhook Server server
kubectl patch deploy -n manba manba-ingress \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"ingress-controller","env":[{"name":"CONTROLLER_ADMISSION_WEBHOOK_LISTEN","value":":8081"}],"volumeMounts":[{"name":"validation-webhook","mountPath":"/admission-webhook"}]}],"volumes":[{"secret":{"secretName":"manba-validation-webhook"},"name":"validation-webhook"}]}}}}'
# configure k8s apiserver to send validations to the webhook
echo "apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: manba-validations
webhooks:
- name: validations.manba.io
  failurePolicy: Fail
  sideEffects: None
  admissionReviewVersions: [\"v1beta1\"]
  rules:
  - apiGroups:
    - configuration.manba.io
    apiVersions:
    - '*'
    operations:
    - CREATE
    - UPDATE
    resources:
    - manbaingresses
    - manbaclusters
  - apiGroups:
    - ''
    apiVersions:
    - 'v1'
    operations:
    - CREATE
    - UPDATE
    resources:
    - secrets
  clientConfig:
    service:
      namespace: manba
      name: manba-validation-webhook
    caBundle: $(cat tls.crt  | base64 | tr -d '\n') " | kubectl apply -f -

