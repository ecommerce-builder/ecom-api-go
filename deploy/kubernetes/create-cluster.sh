#!/bin/bash
gcloud beta container \
  --project "spycameracctv-d48ac" \
  clusters create "cluster-1" \
  --zone "europe-west2-b" \
  --username "admin" \
  --cluster-version "1.9.7-gke.11" \
  --machine-type "g1-small" \
  --image-type "COS" \
  --disk-type "pd-ssd" \
  --disk-size "10" \
  --scopes "https://www.googleapis.com/auth/cloud-platform" \
  --num-nodes "1" \
  --enable-cloud-logging \
  --enable-cloud-monitoring \
  --enable-ip-alias \
  --network "projects/spycameracctv-d48ac/global/networks/default" \
  --subnetwork "projects/spycameracctv-d48ac/regions/europe-west2/subnetworks/default" \
  --default-max-pods-per-node "110" \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing \
  --enable-autoupgrade \
  --enable-autorepair \
  --maintenance-window "03:00"

