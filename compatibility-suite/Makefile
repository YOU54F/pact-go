FEATURES := V1 V2 V3 V4

.PHONY: $(FEATURES)

$(FEATURES):
	go test -test.v -test.run ^TestFeatures$@
