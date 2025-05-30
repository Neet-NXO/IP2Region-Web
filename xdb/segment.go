// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

package xdb

import (
	"fmt"
	"strings"
)

type Segment struct {
	StartIP uint32
	EndIP   uint32
	Region  string
}

func SegmentFrom(seg string) (*Segment, error) {
	var ps = strings.SplitN(strings.TrimSpace(seg), "|", 3)
	if len(ps) != 3 {
		return nil, fmt.Errorf("invalid ip segment `%s`", seg)
	}

	sip, err := IP2Long(ps[0])
	if err != nil {
		return nil, fmt.Errorf("check start ip `%s`: %s", ps[0], err)
	}

	eip, err := IP2Long(ps[1])
	if err != nil {
		return nil, fmt.Errorf("check end ip `%s`: %s", ps[1], err)
	}

	if sip > eip {
		return nil, fmt.Errorf("start ip(%s) should not be greater than end ip(%s)", ps[0], ps[1])
	}

	return &Segment{
		StartIP: sip,
		EndIP:   eip,
		Region:  ps[2],
	}, nil
}

// AfterCheck check the current segment is the one just after the specified one
func (s *Segment) AfterCheck(last *Segment) error {
	if last != nil {
		if last.EndIP+1 != s.StartIP {
			return fmt.Errorf(
				"discontinuous data segment: last.eip+1(%d) != seg.sip(%d, %s)",
				last.EndIP+1, s.StartIP, s.Region,
			)
		}
	}

	return nil
}

// Split the segment based on the pre-two bytes
func (s *Segment) Split() []*Segment {
	// 1, split the segment with the first byte
	var tList []*Segment
	var sByte1, eByte1 = (s.StartIP >> 24) & 0xFF, (s.EndIP >> 24) & 0xFF
	var nSip = s.StartIP // nSip tracks the start of the current effective block for the first byte split

	for i := sByte1; i <= eByte1; i++ {
		// sip for this first-byte-block starts at i.0.0.0, but respecting original nSip lower bytes
		// e.g. if nSip is 1.2.3.4, and i is 1, then sip is 1.2.3.4
		// if nSip is 1.2.3.4, and i is 2, then sip is 2.0.0.0 (because nSip will be updated)
		sip1 := (i << 24) | (nSip & 0x00FFFFFF)
		eip1 := (i << 24) | 0x00FFFFFF // Ends at i.255.255.255

		// Adjust sip1: it cannot be smaller than the original segment's StartIP
		if sip1 < s.StartIP {
			sip1 = s.StartIP
		}

		// Adjust eip1: it cannot be larger than the original segment's EndIP
		if eip1 > s.EndIP {
			eip1 = s.EndIP
		}

		if sip1 <= eip1 { // Only add if the range is valid
			tSeg := &Segment{
				StartIP: sip1,
				EndIP:   eip1,
				// Region is NOT set here, will be set in the second phase from original s.Region
			}
			tList = append(tList, tSeg)
		}

		// Prepare nSip for the next iteration of the first byte (i+1)
		// If eip1 (i.255.255.255, possibly capped by s.EndIP) is less than s.EndIP,
		// it means the original segment continues into the next first-byte block ((i+1).x.x.x).
		// So, the next nSip should start at (i+1).0.0.0.
		if eip1 < s.EndIP {
			nSip = (i + 1) << 24
		} // else, original segment ends within this first-byte block, nSip doesn't need to advance past s.EndIP for this logic.
	}

	// 2, split the segments from tList with the second byte
	var segList []*Segment
	for _, segFromTList := range tList { // segFromTList is like A.B.C.D | A.Y.Z.W (Region not set yet)
		base := segFromTList.StartIP & 0xFF000000 // First byte, e.g., 223.0.0.0
		// nSip2 tracks the start for the second byte split, relative to segFromTList.StartIP
		nSip2 := segFromTList.StartIP

		sb2, eb2 := (segFromTList.StartIP>>16)&0xFF, (segFromTList.EndIP>>16)&0xFF // Second byte range for this tList segment

		for j := sb2; j <= eb2; j++ { // Iterate over second byte values
			// Theoretical sub-block for this second byte j: A.j.0.0 to A.j.255.255
			// Use nSip2 to respect the lower two bytes of segFromTList.StartIP for the first 'j'
			subBlockSip := base | (uint32(j) << 16) | (nSip2 & 0x0000FFFF)
			subBlockEip := base | (uint32(j) << 16) | 0x0000FFFF

			// Actual SIP: Cannot be less than the original segment's StartIP (s.StartIP)
			// AND cannot be less than the current tList segment's StartIP (segFromTList.StartIP)
			actualSip := subBlockSip
			if actualSip < s.StartIP {
				actualSip = s.StartIP
			}
			if actualSip < segFromTList.StartIP {
				actualSip = segFromTList.StartIP
			}

			// Actual EIP: Cannot be greater than the original segment's EndIP (s.EndIP)
			// AND cannot be greater than the current tList segment's EndIP (segFromTList.EndIP)
			actualEip := subBlockEip
			if actualEip > s.EndIP {
				actualEip = s.EndIP
			}
			if actualEip > segFromTList.EndIP {
				actualEip = segFromTList.EndIP
			}

			if actualSip <= actualEip { // Only add if the resulting range is valid
				smallSeg := &Segment{
					StartIP: actualSip,
					EndIP:   actualEip,
					Region:  s.Region, // Use region from the original segment 's'
				}
				segList = append(segList, smallSeg)
			}

			// Prepare nSip2 for the next iteration of the second byte (j+1)
			if actualEip < segFromTList.EndIP && actualEip < s.EndIP {
				nSip2 = base | ((uint32(j) + 1) << 16)
			}
		}
	}
	return segList
}

func (s *Segment) String() string {
	return fmt.Sprintf("%s|%s|%s", Long2IP(s.StartIP), Long2IP(s.EndIP), s.Region)
}
