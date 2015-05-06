package local

import (
	"testing"

	clientmodel "github.com/prometheus/client_golang/model"
)

var (
	// cm11, cm12, cm13 are colliding with fp1.
	// cm21, cm22 are colliding with fp2.
	// cm31, cm32 are colliding with fp3, which is below maxMappedFP.
	// Note that fingerprints are set and not actually calculated.
	// The collision detection is independent from the actually used
	// fingerprinting algorithm.
	fp1  = clientmodel.Fingerprint(maxMappedFP + 1)
	fp2  = clientmodel.Fingerprint(maxMappedFP + 2)
	fp3  = clientmodel.Fingerprint(1)
	cm11 = clientmodel.Metric{
		"foo":   "bar",
		"dings": "bumms",
	}
	cm12 = clientmodel.Metric{
		"bar": "foo",
	}
	cm13 = clientmodel.Metric{
		"foo": "bar",
	}
	cm21 = clientmodel.Metric{
		"foo":   "bumms",
		"dings": "bar",
	}
	cm22 = clientmodel.Metric{
		"dings": "foo",
		"bar":   "bumms",
	}
	cm31 = clientmodel.Metric{
		"bumms": "dings",
	}
	cm32 = clientmodel.Metric{
		"bumms": "dings",
		"bar":   "foo",
	}
)

func TestFPMapper(t *testing.T) {
	sm := newSeriesMap()

	p, closer := newTestPersistence(t, 1)
	defer closer.Close()

	mapper, err := newFPMapper(sm, p)
	if err != nil {
		t.Fatal(err)
	}

	// Everything is empty, resolving a FP should do nothing.
	gotFP, err := mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// cm11 is in sm. Adding cm11 should do nothing. Mapping cm12 should resolve
	// the collision.
	sm.put(fp1, &memorySeries{metric: cm11})
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// The mapped cm12 is added to sm, too. That should not change the outcome.
	sm.put(clientmodel.Fingerprint(1), &memorySeries{metric: cm12})
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Now map cm13, should reproducibly result in the next mapped FP.
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Add cm13 to sm. Should not change anything.
	sm.put(clientmodel.Fingerprint(2), &memorySeries{metric: cm13})
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Now add cm21 and cm22 in the same way, checking the mapped FPs.
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	sm.put(fp2, &memorySeries{metric: cm21})
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(3); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	sm.put(clientmodel.Fingerprint(3), &memorySeries{metric: cm22})
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(3); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Map cm31, resulting in a mapping straight away.
	gotFP, err = mapper.mapFP(fp3, cm31)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(4); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	sm.put(clientmodel.Fingerprint(4), &memorySeries{metric: cm31})

	// Map cm32, which is now mapped for two reasons...
	gotFP, err = mapper.mapFP(fp3, cm32)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(5); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	sm.put(clientmodel.Fingerprint(5), &memorySeries{metric: cm32})

	// Now check ALL the mappings, just to be sure.
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(3); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm31)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(4); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm32)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(5); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Remove all the fingerprints from sm, which should change nothing, as
	// the existing mappings stay and should be detected.
	sm.del(fp1)
	sm.del(fp2)
	sm.del(fp3)
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(3); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm31)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(4); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm32)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(5); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// Load the mapper anew from disk and then check all the mappings again
	// to make sure all changes have made it to disk.
	mapper, err = newFPMapper(sm, p)
	if err != nil {
		t.Fatal(err)
	}
	gotFP, err = mapper.mapFP(fp1, cm11)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(1); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp1, cm13)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(2); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(3); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm31)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(4); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp3, cm32)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(5); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// To make sure that the mapping layer is not queried if the FP is found
	// either in sm or in the archive, now put fp1 with cm12 in sm and fp2
	// with cm22 into archive (which will never happen in practice as only
	// mapped FPs are put into sm and the archive.
	sm.put(fp1, &memorySeries{metric: cm12})
	p.archiveMetric(fp2, cm22, 0, 0)
	gotFP, err = mapper.mapFP(fp1, cm12)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp1; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}
	gotFP, err = mapper.mapFP(fp2, cm22)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := fp2; gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

	// If we now map cm21, we should get a mapping as the collision with the
	// archived metric is detected.
	gotFP, err = mapper.mapFP(fp2, cm21)
	if err != nil {
		t.Fatal(err)
	}
	if wantFP := clientmodel.Fingerprint(6); gotFP != wantFP {
		t.Errorf("got fingerprint %v, want fingerprint %v", gotFP, wantFP)
	}

}
