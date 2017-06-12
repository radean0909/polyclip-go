package polyclip_test

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	polyclip "github.com/ctessum/polyclip-go"
)

type sorter polyclip.Polygon

func (s sorter) Len() int      { return len(s) }
func (s sorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sorter) Less(i, j int) bool {
	if len(s[i]) != len(s[j]) {
		return len(s[i]) < len(s[j])
	}
	for k := range s[i] {
		pi, pj := s[i][k], s[j][k]
		if pi.X != pj.X {
			return pi.X < pj.X
		}
		if pi.Y != pj.Y {
			return pi.Y < pj.Y
		}
	}
	return false
}

// basic normalization just for tests; to be improved if needed
func normalize(poly polyclip.Polygon) polyclip.Polygon {
	for i, c := range poly {
		if len(c) == 0 {
			continue
		}

		// find bottom-most of leftmost points, to have fixed anchor
		min := 0
		for j, p := range c {
			if p.X < c[min].X || p.X == c[min].X && p.Y < c[min].Y {
				min = j
			}
		}

		// rotate points to make sure min is first
		poly[i] = append(c[min:], c[:min]...)
	}

	sort.Sort(sorter(poly))
	return poly
}

func dump(poly polyclip.Polygon) string {
	return fmt.Sprintf("%v", normalize(poly))
}

func TestBug3(t *testing.T) {
	cases := []struct{ subject, clipping, result polyclip.Polygon }{
		// original reported github issue #3
		{
			subject: polyclip.Polygon{{{1, 1}, {1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{
				{{2, 1}, {2, 2}, {3, 2}, {3, 1}},
				{{1, 2}, {1, 3}, {2, 3}, {2, 2}},
				{{2, 2}, {2, 3}, {3, 3}, {3, 2}}},
			result: polyclip.Polygon{{
				{1, 1}, {2, 1}, {3, 1},
				{3, 2}, {3, 3},
				{2, 3}, {1, 3},
				{1, 2}}},
		},
		// simplified variant of issue #3, for easier debugging
		{
			subject: polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{
				{{2, 1}, {2, 2}, {3, 2}},
				{{1, 2}, {2, 3}, {2, 2}},
				{{2, 2}, {2, 3}, {3, 2}}},
			result: polyclip.Polygon{{{1, 2}, {2, 3}, {3, 2}, {2, 1}}},
		},
		{
			subject: polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{
				{{1, 2}, {2, 3}, {2, 2}},
				{{2, 2}, {2, 3}, {3, 2}}},
			result: polyclip.Polygon{{{1, 2}, {2, 3}, {3, 2}, {2, 2}, {2, 1}}},
		},
		// another variation, now with single degenerated curve
		{
			subject: polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{
				{{1, 2}, {2, 3}, {2, 2}, {2, 3}, {3, 2}}},
			result: polyclip.Polygon{{{1, 2}, {2, 3}, {3, 2}, {2, 2}, {2, 1}}},
		},
		{
			subject: polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{
				{{2, 1}, {2, 2}, {2, 3}, {3, 2}},
				{{1, 2}, {2, 3}, {2, 2}}},
			result: polyclip.Polygon{{{1, 2}, {2, 3}, {3, 2}, {2, 1}}},
		},
		// "union" with effectively empty polygon (wholly self-intersecting)
		{
			subject:  polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
			clipping: polyclip.Polygon{{{1, 2}, {2, 2}, {2, 3}, {1, 2}, {2, 2}, {2, 3}}},
			result:   polyclip.Polygon{{{1, 2}, {2, 2}, {2, 1}}},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result := dump(c.subject.Construct(polyclip.UNION, c.clipping))
			if result != dump(c.result) {
				t.Errorf("case UNION:\nsubject:  %v\nclipping: %v\nexpected: %v\ngot:      %v",
					c.subject, c.clipping, c.result, result)
			}
		})
	}
}

func TestNonReductiveSegmentDivisions(t *testing.T) {
	if testing.Short() {
		return
	}

	cases := []struct{ subject, clipping polyclip.Polygon }{
		{
			// original reported github issue #4, resulting in infinite loop
			subject: polyclip.Polygon{{
				{X: 1.427255375e+06, Y: -2.3283064365386963e-10},
				{X: 1.4271285e+06, Y: 134.7111358642578},
				{X: 1.427109e+06, Y: 178.30108642578125}}},
			clipping: polyclip.Polygon{{
				{X: 1.416e+06, Y: -12000},
				{X: 1.428e+06, Y: -12000},
				{X: 1.428e+06, Y: 0},
				{X: 1.416e+06, Y: 0},
				{X: 1.416e+06, Y: -12000}}},
		},
		// Test cases from https://github.com/ctessum/polyclip-go/blob/master/bugs_test.go
		{
			subject: polyclip.Polygon{{
				{X: 1.7714672107465276e+06, Y: -102506.68254093888},
				{X: 1.7713768917571804e+06, Y: -102000.75485953009},
				{X: 1.7717109214841307e+06, Y: -101912.19625031832}}},
			clipping: polyclip.Polygon{{
				{X: 1.7714593229229522e+06, Y: -102470.35230830211},
				{X: 1.7714672107465276e+06, Y: -102506.68254093867},
				{X: 1.771439738086082e+06, Y: -102512.92027456204}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: -1.8280000000000012e+06, Y: -492999.99999999953},
				{X: -1.8289999999999995e+06, Y: -494000.0000000006},
				{X: -1.828e+06, Y: -493999.9999999991},
				{X: -1.8280000000000012e+06, Y: -492999.99999999953}}},
			clipping: polyclip.Polygon{{
				{X: -1.8280000000000005e+06, Y: -495999.99999999977},
				{X: -1.8280000000000007e+06, Y: -492000.0000000014},
				{X: -1.8240000000000007e+06, Y: -492000.0000000014},
				{X: -1.8280000000000005e+06, Y: -495999.99999999977}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: -2.0199999999999988e+06, Y: -394999.99999999825},
				{X: -2.0199999999999988e+06, Y: -392000.0000000009},
				{X: -2.0240000000000012e+06, Y: -395999.9999999993},
				{X: -2.0199999999999988e+06, Y: -394999.99999999825}}},
			clipping: polyclip.Polygon{{
				{X: -2.0199999999999988e+06, Y: -394999.99999999825},
				{X: -2.020000000000001e+06, Y: -394000.0000000001},
				{X: -2.0190000000000005e+06, Y: -394999.9999999997},
				{X: -2.0199999999999988e+06, Y: -394999.99999999825}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: -47999.99999999992, Y: -23999.999999998756},
				{X: 0, Y: -24000.00000000017},
				{X: 0, Y: 24000.00000000017},
				{X: -48000.00000000014, Y: 24000.00000000017},
				{X: -47999.99999999992, Y: -23999.999999998756}}},
			clipping: polyclip.Polygon{{
				{X: -48000, Y: -24000},
				{X: 0, Y: -24000},
				{X: 0, Y: 24000},
				{X: -48000, Y: 24000},
				{X: -48000, Y: -24000}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: -2.137000000000001e+06, Y: -122000.00000000093},
				{X: -2.1360000000000005e+06, Y: -121999.99999999907},
				{X: -2.1360000000000014e+06, Y: -121000.00000000186}}},
			clipping: polyclip.Polygon{{
				{X: -2.1120000000000005e+06, Y: -120000},
				{X: -2.136000000000001e+06, Y: -120000.00000000093},
				{X: -2.1360000000000005e+06, Y: -144000}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: 1.556e+06, Y: -1.139999999999999e+06},
				{X: 1.5600000000000002e+06, Y: -1.140000000000001e+06},
				{X: 1.56e+06, Y: -1.136000000000001e+06}}},
			clipping: polyclip.Polygon{{
				{X: 1.56e+06, Y: -1.127999999999999e+06},
				{X: 1.5600000000000002e+06, Y: -1.151999999999999e+06}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: 1.0958876176594219e+06, Y: -567467.5197556159},
				{X: 1.0956330600760083e+06, Y: -567223.72588934},
				{X: 1.0958876176594219e+06, Y: -567467.5197556159}}},
			clipping: polyclip.Polygon{{
				{X: 1.0953516248896217e+06, Y: -564135.1861293605},
				{X: 1.0959085007300845e+06, Y: -568241.1879245406},
				{X: 1.0955136237022132e+06, Y: -581389.3748769956}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: 608000, Y: -113151.36476426799},
				{X: 608000, Y: -114660.04962779157},
				{X: 612000, Y: -115414.39205955336},
				{X: 1.616e+06, Y: -300000},
				{X: 1.608e+06, Y: -303245.6575682382},
				{X: 0, Y: 0}}},
			clipping: polyclip.Polygon{{
				{X: 1.612e+06, Y: -296000}}},
		},
		{
			subject: polyclip.Polygon{{
				{X: 1.1458356382266793e+06, Y: -251939.4635597784},
				{X: 1.1460824662209095e+06, Y: -251687.86194535438},
				{X: 1.1458356382266793e+06, Y: -251939.4635597784}}},
			clipping: polyclip.Polygon{{
				{X: 1.1486683769211173e+06, Y: -251759.06331944838},
				{X: 1.1468807511323579e+06, Y: -251379.90576799586},
				{X: 1.1457914974731328e+06, Y: -251816.31287551578}}},
		},
		{
			// From https://github.com/ctessum/polyclip-go/commit/6614925d6d7087b7afcd4c55571554f67efd2ec3
			subject: polyclip.Polygon{{
				{X: 426694.6365274183, Y: -668547.1611580737},
				{X: 426714.57523030025, Y: -668548.9238652373},
				{X: 426745.39648089616, Y: -668550.4651249861}}},
			clipping: polyclip.Polygon{{
				{X: 426714.5752302991, Y: -668548.9238652373},
				{X: 426744.63718662335, Y: -668550.0591896093},
				{X: 426745.3964821229, Y: -668550.4652243527}}},
		},
		{
			// Produces invalid divisions that would otherwise continually generate new segments.
			subject: polyclip.Polygon{{
				{X: 99.67054939325573, Y: 23.50752393246498},
				{X: 99.88993946188153, Y: 20.999883973365655},
				{X: 100.01468418889, Y: 20.53433031419374}}},
			clipping: polyclip.Polygon{{
				{X: 100.15374164547939, Y: 20.015360821030836},
				{X: 95.64222842284941, Y: 36.85255738690467},
				{X: 100.15374164547939, Y: -14.714274712355238}}},
		},
	}

	for _, c := range cases {
		const rotations = 360
		// Test multiple rotations of each case to catch any orientation assumptions.
		for i := 0; i < rotations; i++ {
			angle := 2 * math.Pi * float64(i) / float64(rotations)
			subject := rotate(c.subject, angle)
			clipping := rotate(c.clipping, angle)

			for _, op := range []polyclip.Op{polyclip.UNION, polyclip.INTERSECTION, polyclip.DIFFERENCE} {
				ch := make(chan polyclip.Polygon)
				go func() {
					ch <- subject.Construct(op, clipping)
				}()

				select {
				case <-ch:
					// check that we get a result in finite time
				case <-time.After(1 * time.Second):
					// panicking in attempt to get full stacktrace
					panic(fmt.Sprintf("case %v:\nsubject:  %v\nclipping: %v\ntimed out.", op, subject, clipping))
				}
			}
		}
	}
}

func rotate(p polyclip.Polygon, radians float64) polyclip.Polygon {
	result := p.Clone()
	for i, contour := range p {
		result[i] = make(polyclip.Contour, len(contour))
		for j, point := range contour {
			result[i][j] = polyclip.Point{
				X: point.X*math.Cos(radians) - point.Y*math.Sin(radians),
				Y: point.Y*math.Cos(radians) + point.X*math.Sin(radians),
			}
		}
	}
	return result
}

func TestBug5(t *testing.T) {
	rect := polyclip.Polygon{{{24, 7}, {36, 7}, {36, 23}, {24, 23}}}
	circle := polyclip.Polygon{{{24, 7}, {24.83622770614123, 7.043824837053814}, {25.66329352654208, 7.174819194129555}, {26.472135954999587, 7.391547869638773}, {27.253893144606412, 7.691636338859195}, {28.00000000000001, 8.071796769724493}, {28.702282018339798, 8.527864045000424}, {29.35304485087088, 9.054841396180851}, {29.94515860381917, 9.646955149129141}, {30.472135954999597, 10.297717981660224}, {30.92820323027553, 11.00000000000001}, {31.308363661140827, 11.746106855393611}, {31.60845213036125, 12.527864045000435}, {31.825180805870467, 13.33670647345794}, {31.95617516294621, 14.16377229385879}, {32.00000000000002, 15.00000000000002}, {31.95617516294621, 15.83622770614125}, {31.825180805870467, 16.6632935265421}, {31.60845213036125, 17.472135954999604}, {31.308363661140827, 18.25389314460643}, {30.92820323027553, 19.00000000000003}, {30.472135954999597, 19.702282018339815}, {29.94515860381917, 20.353044850870898}, {29.35304485087088, 20.945158603819188}, {28.702282018339798, 21.472135954999615}, {28.00000000000001, 21.928203230275546}, {27.253893144606412, 22.308363661140845}, {26.472135954999587, 22.608452130361268}, {25.66329352654208, 22.825180805870485}, {24.83622770614123, 22.956175162946227}, {24, 23.00000000000004}, {23.16377229385877, 22.956175162946227}, {22.33670647345792, 22.825180805870485}, {21.527864045000413, 22.608452130361268}, {20.746106855393588, 22.308363661140845}, {19.99999999999999, 21.928203230275546}, {19.297717981660202, 21.472135954999615}, {18.64695514912912, 20.945158603819188}, {18.05484139618083, 20.353044850870898}, {17.527864045000403, 19.702282018339815}, {17.07179676972447, 19.00000000000003}, {16.691636338859173, 18.25389314460643}, {16.39154786963875, 17.472135954999604}, {16.174819194129533, 16.6632935265421}, {16.04382483705379, 15.83622770614125}, {15.999999999999977, 15.00000000000002}, {16.04382483705379, 14.16377229385879}, {16.174819194129533, 13.33670647345794}, {16.39154786963875, 12.527864045000435}, {16.691636338859173, 11.746106855393611}, {17.07179676972447, 11.00000000000001}, {17.527864045000403, 10.297717981660224}, {18.05484139618083, 9.646955149129141}, {18.64695514912912, 9.054841396180851}, {19.297717981660202, 8.527864045000424}, {19.99999999999999, 8.071796769724493}, {20.746106855393588, 7.691636338859194}, {21.527864045000413, 7.391547869638772}, {22.33670647345792, 7.1748191941295545}, {23.16377229385877, 7.043824837053813}}}

	expected := []struct {
		op     polyclip.Op
		result polyclip.Polygon
	}{
		{
			polyclip.UNION,
			polyclip.Polygon{{{36, 23}, {36, 7}, {24, 7}, {23.16377229385877, 7.043824837053813}, {22.33670647345792, 7.1748191941295545}, {21.527864045000413, 7.391547869638772}, {20.746106855393588, 7.691636338859194}, {19.99999999999999, 8.071796769724493}, {19.297717981660202, 8.527864045000424}, {18.64695514912912, 9.054841396180851}, {18.05484139618083, 9.646955149129141}, {17.527864045000403, 10.297717981660224}, {17.07179676972447, 11.00000000000001}, {16.691636338859173, 11.746106855393611}, {16.39154786963875, 12.527864045000435}, {16.174819194129533, 13.33670647345794}, {16.04382483705379, 14.16377229385879}, {15.999999999999977, 15.00000000000002}, {16.04382483705379, 15.83622770614125}, {16.174819194129533, 16.6632935265421}, {16.39154786963875, 17.472135954999604}, {16.691636338859173, 18.25389314460643}, {17.07179676972447, 19.00000000000003}, {17.527864045000403, 19.702282018339815}, {18.05484139618083, 20.353044850870898}, {18.64695514912912, 20.945158603819188}, {19.297717981660202, 21.472135954999615}, {19.99999999999999, 21.928203230275546}, {20.746106855393588, 22.308363661140845}, {21.527864045000413, 22.608452130361268}, {22.33670647345792, 22.825180805870485}, {23.16377229385877, 22.956175162946227}, {24, 23.00000000000004}, {24.000000000000746, 23}}},
		},
		{
			polyclip.INTERSECTION,
			polyclip.Polygon{{{31.95617516294621, 15.83622770614125}, {31.825180805870467, 16.6632935265421}, {31.60845213036125, 17.472135954999604}, {31.308363661140827, 18.25389314460643}, {30.92820323027553, 19.00000000000003}, {30.472135954999597, 19.702282018339815}, {29.94515860381917, 20.353044850870898}, {29.35304485087088, 20.945158603819188}, {28.702282018339798, 21.472135954999615}, {28.00000000000001, 21.928203230275546}, {27.253893144606412, 22.308363661140845}, {26.472135954999587, 22.608452130361268}, {25.66329352654208, 22.825180805870485}, {24.83622770614123, 22.956175162946227}, {24.000000000000746, 23}, {24, 23}, {24, 7}, {24.83622770614123, 7.043824837053814}, {25.66329352654208, 7.174819194129555}, {26.472135954999587, 7.391547869638773}, {27.253893144606412, 7.691636338859195}, {28.00000000000001, 8.071796769724493}, {28.702282018339798, 8.527864045000424}, {29.35304485087088, 9.054841396180851}, {29.94515860381917, 9.646955149129141}, {30.472135954999597, 10.297717981660224}, {30.92820323027553, 11.00000000000001}, {31.308363661140827, 11.746106855393611}, {31.60845213036125, 12.527864045000435}, {31.825180805870467, 13.33670647345794}, {31.95617516294621, 14.16377229385879}, {32.00000000000002, 15.00000000000002}}},
		},
		{
			polyclip.DIFFERENCE,
			polyclip.Polygon{{{24.000000000000746, 23}, {24.83622770614123, 22.956175162946227}, {25.66329352654208, 22.825180805870485}, {26.472135954999587, 22.608452130361268}, {27.253893144606412, 22.308363661140845}, {28.00000000000001, 21.928203230275546}, {28.702282018339798, 21.472135954999615}, {29.35304485087088, 20.945158603819188}, {29.94515860381917, 20.353044850870898}, {30.472135954999597, 19.702282018339815}, {30.92820323027553, 19.00000000000003}, {31.308363661140827, 18.25389314460643}, {31.60845213036125, 17.472135954999604}, {31.825180805870467, 16.6632935265421}, {31.95617516294621, 15.83622770614125}, {32.00000000000002, 15.00000000000002}, {31.95617516294621, 14.16377229385879}, {31.825180805870467, 13.33670647345794}, {31.60845213036125, 12.527864045000435}, {31.308363661140827, 11.746106855393611}, {30.92820323027553, 11.00000000000001}, {30.472135954999597, 10.297717981660224}, {29.94515860381917, 9.646955149129141}, {29.35304485087088, 9.054841396180851}, {28.702282018339798, 8.527864045000424}, {28.00000000000001, 8.071796769724493}, {27.253893144606412, 7.691636338859195}, {26.472135954999587, 7.391547869638773}, {25.66329352654208, 7.174819194129555}, {24.83622770614123, 7.043824837053814}, {24, 7}, {36, 7}, {36, 23}}},
		},
		{
			polyclip.XOR,
			polyclip.Polygon{
				{{24.000000000000746, 23}, {24, 23}, {24, 7}, {23.16377229385877, 7.043824837053813}, {22.33670647345792, 7.1748191941295545}, {21.527864045000413, 7.391547869638772}, {20.746106855393588, 7.691636338859194}, {19.99999999999999, 8.071796769724493}, {19.297717981660202, 8.527864045000424}, {18.64695514912912, 9.054841396180851}, {18.05484139618083, 9.646955149129141}, {17.527864045000403, 10.297717981660224}, {17.07179676972447, 11.00000000000001}, {16.691636338859173, 11.746106855393611}, {16.39154786963875, 12.527864045000435}, {16.174819194129533, 13.33670647345794}, {16.04382483705379, 14.16377229385879}, {15.999999999999977, 15.00000000000002}, {16.04382483705379, 15.83622770614125}, {16.174819194129533, 16.6632935265421}, {16.39154786963875, 17.472135954999604}, {16.691636338859173, 18.25389314460643}, {17.07179676972447, 19.00000000000003}, {17.527864045000403, 19.702282018339815}, {18.05484139618083, 20.353044850870898}, {18.64695514912912, 20.945158603819188}, {19.297717981660202, 21.472135954999615}, {19.99999999999999, 21.928203230275546}, {20.746106855393588, 22.308363661140845}, {21.527864045000413, 22.608452130361268}, {22.33670647345792, 22.825180805870485}, {23.16377229385877, 22.956175162946227}, {24, 23.00000000000004}},
				{{24.000000000000746, 23}, {24.83622770614123, 22.956175162946227}, {25.66329352654208, 22.825180805870485}, {26.472135954999587, 22.608452130361268}, {27.253893144606412, 22.308363661140845}, {28.00000000000001, 21.928203230275546}, {28.702282018339798, 21.472135954999615}, {29.35304485087088, 20.945158603819188}, {29.94515860381917, 20.353044850870898}, {30.472135954999597, 19.702282018339815}, {30.92820323027553, 19.00000000000003}, {31.308363661140827, 18.25389314460643}, {31.60845213036125, 17.472135954999604}, {31.825180805870467, 16.6632935265421}, {31.95617516294621, 15.83622770614125}, {32.00000000000002, 15.00000000000002}, {31.95617516294621, 14.16377229385879}, {31.825180805870467, 13.33670647345794}, {31.60845213036125, 12.527864045000435}, {31.308363661140827, 11.746106855393611}, {30.92820323027553, 11.00000000000001}, {30.472135954999597, 10.297717981660224}, {29.94515860381917, 9.646955149129141}, {29.35304485087088, 9.054841396180851}, {28.702282018339798, 8.527864045000424}, {28.00000000000001, 8.071796769724493}, {27.253893144606412, 7.691636338859195}, {26.472135954999587, 7.391547869638773}, {25.66329352654208, 7.174819194129555}, {24.83622770614123, 7.043824837053814}, {24, 7}, {36, 7}, {36, 23}},
			},
		},
	}

	for _, e := range expected {
		result := rect.Construct(e.op, circle)
		if dump(result) != dump(e.result) {
			t.Errorf("case %d expected:\n%v\ngot:\n%v", e.op, dump(e.result), dump(result))
		}
	}
}

func TestSelfIntersect(t *testing.T) {
	rect1 := polyclip.Polygon{{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, {{1, 0}, {2, 0}, {2, 1}, {1, 1}}}
	rect2 := polyclip.Polygon{{{0, 0.25}, {3, 0.25}, {3, 0.75}, {0, 0.75}}}

	expected := []struct {
		name   string
		op     polyclip.Op
		result polyclip.Polygon
	}{
		{
			"union",
			polyclip.UNION,
			polyclip.Polygon{{{0, 0}, {1, 0}, {2, 0}, {2, 0.25}, {3, 0.25}, {3, 0.75}, {2, 0.75}, {2, 1}, {1, 1}, {0, 1}, {0, 0.75}, {0, 0.25}}},
		},
		{
			"intersection",
			polyclip.INTERSECTION,
			polyclip.Polygon{{{0, 0.25}, {2, 0.25}, {2, 0.75}, {0, 0.75}}},
		},
		{
			"difference",
			polyclip.DIFFERENCE,
			polyclip.Polygon{{{0, 0}, {1, 0}, {2, 0}, {2, 0.25}, {0, 0.25}}, {{0, 0.75}, {2, 0.75}, {2, 1}, {1, 1}, {0, 1}}},
		},
		{
			"xor",
			polyclip.XOR,
			// TODO: This one is a little weird.  It probably shouldn't be self-intersecting.
			polyclip.Polygon{{{0, 0}, {1, 0}, {2, 0}, {2, 0.25}, {0, 0.25}}, {{0, 0.75}, {2, 0.75}, {2, 0.25}, {3, 0.25}, {3, 0.75}, {2, 0.75}, {2, 1}, {1, 1}, {0, 1}}},
		},
	}

	for _, e := range expected {
		t.Run(e.name, func(t *testing.T) {
			result := rect1.Construct(e.op, rect2)
			if dump(result) != dump(e.result) {
				t.Errorf("case %d expected:\n%v\ngot:\n%v", e.op, dump(e.result), dump(result))
			}
		})
	}
}
