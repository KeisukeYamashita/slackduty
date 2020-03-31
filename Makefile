.PHONY: build
build:
	docker build . -t slackduty

.PHONY: run
run: build
	docker run -ti -v ~/.slackduty/config.yml:/root/.slackduty/config.yml \
		-e SLACKDUTY_CONFIG=${SLACKDUTY_CONFIG} \
		-e SLACKDUTY_SLACK_API_KEY=${SLACKDUTY_SLACK_API_KEY} \
		-e SLACKDUTY_PAGERDUTY_API_KEY=${SLACKDUTY_PAGERDUTY_API_KEY} slackduty

.PHONY: deploy-k8s
deploy-k8s:
	kubectl create configmap slackduty-configmap --from-file ${SLACKDUTY_CONFIG}
	kubectl create secret generic slackduty-api-key --from-literal=slack-api-key=${SLACKDUTY_SLACK_API_KEY} --from-literal=pagerduty-api-key=${SLACKDUTY_PAGERDUTY_API_KEY} 
	kubectl apply -f k8s/cronjob.yml