interface StarHR {
	hip: number;
	Amag: number;
	BV: number;
}

/*
0. Catalog ( Catalog_Name ) - Catalogue (H=Hipparcos) 
1. HIP ( HIP_Number ) - Identifier (HIP number) 
2. Proxy ( Prox_10asec ) - Proximity flag 
3. RAhms ( RA ) - RA in h m s, ICRS (J1991.25) 
4. DEdms ( Dec ) - Dec in deg ' ", ICRS (J1991.25) 
5. Vmag ( Vmag ) - Magnitude in Johnson V 
6. VarFlag ( Var_Flag ) - Coarse variability flag 
7. r_Vmag ( Vmag_Source ) - Source of magnitude 
8. RAdeg ( RA_Deg ) - RA in degrees (ICRS, Epoch-J1991.25) 
9. DEdeg ( Dec_Deg ) - Dec in degrees (ICRS, Epoch-J1991.25) 
10. AstroRef ( Astrom_Ref_Dbl ) - Reference flag for astrometry 
11. Plx ( Parallax ) - Trigonometric parallax 
12. pmRA ( pm_RA ) - Proper motion in RA 
13. pmDE ( pm_Dec ) - Proper motion in Dec 
14. e_RAdeg ( RA_Error ) - Standard error in RA*cos(Dec_Deg) 
15. e_DEdeg ( Dec_Error ) - Standard error in Dec_Deg 
16. e_Plx ( Parallax_Error ) - Standard error in Parallax 
17. e_pmRA ( pm_RA_Error ) - Standard error in pmRA 
18. e_pmDE ( pm_Dec_Error ) - Standard error in pmDE 
19. DE:RA ( Crl_Dec_RA ) - (DE over RA)xCos(delta) 
20. Plx:RA ( Crl_Plx_RA ) - (Plx over RA)xCos(delta) 
21. Plx:DE ( Crl_Plx_Dec ) - (Plx over DE) 
22. pmRA:RA ( Crl_pmRA_RA ) - (pmRA over RA)xCos(delta) 
23. pmRA:DE ( Crl_pmRA_Dec ) - (pmRA over DE) 
24. pmRA:Plx ( Crl_pmRA_Plx ) - (pmRA over Plx) 
25. pmDE:RA ( Crl_pmDec_RA ) - (pmDE over RA)xCos(delta) 
26. pmDE:DE ( Crl_pmDec_Dec ) - (pmDE over DE) 
27. pmDE:Plx ( Crl_pmDec_Plx ) - (pmDE over Plx) 
28. pmDE:pmRA ( Crl_pmDec_pmRA ) - (pmDE over pmRA) 
29. F1 ( Reject_Percent ) - Percentage of rejected data 
30. F2 ( Quality_Fit ) - Goodness-of-fit parameter 
31. --- ( HIP_Number_repeat ) - HIP number (repetition) 
32. BTmag ( BT_Mag ) - Mean BT magnitude 
33. e_BTmag ( BT_Mag_Error ) - Standard error on BTmag 
34. VTmag ( VT_Mag ) - Mean VT magnitude 
35. e_VTmag ( VT_Mag_Error ) - Standard error on VTmag 
36. m_BTmag ( BT_Mag_Ref_Dbl ) - Reference flag for BT and VTmag 
37. B-V ( BV_Color ) - Johnson BV colour 
38. e_B-V ( BV_Color_Error ) - Standard error on BV 
39. r_B-V ( BV_Mag_Source ) - Source of BV from Ground or Tycho 
40. V-I ( VI_Color ) - Colour index in Cousins' system 
41. e_V-I ( VI_Color_Error ) - Standard error on VI 
42. r_V-I ( VI_Color_Source ) - Source of VI 
43. CombMag ( Mag_Ref_Dbl ) - Flag for combined Vmag, BV, VI 
44. Hpmag ( Hip_Mag ) - Median magnitude in Hipparcos system 
45. e_Hpmag ( Hip_Mag_Error ) - Standard error on Hpmag 
46. Hpscat ( Scat_Hip_Mag ) - Scatter of Hpmag 
47. o_Hpmag ( N_Obs_Hip_Mag ) - Number of observations for Hpmag 
48. m_Hpmag ( Hip_Mag_Ref_Dbl ) - Reference flag for Hpmag 
49. Hpmax ( Hip_Mag_Max ) - Hpmag at maximum (5th percentile) 
50. HPmin ( Hip_Mag_Min ) - Hpmag at minimum (95th percentile) 
51. Period ( Var_Period ) - Variability period (days) 
52. HvarType ( Hip_Var_Type ) - Variability type 
53. moreVar ( Var_Data_Annex ) - Additional data about variability 
54. morePhoto ( Var_Curv_Annex ) - Light curve Annex 
55. CCDM ( CCDM_Id ) - CCDM identifier 
56. n_CCDM ( CCDM_History ) - Historical status flag 
57. Nsys ( CCDM_N_Entries ) - Number of entries with same CCDM 
58. Ncomp ( CCDM_N_Comp ) - Number of components in this entry 
59. MultFlag ( Dbl_Mult_Annex ) - Double and or Multiple Systems flag 
60. Source ( Astrom_Mult_Source ) - Astrometric source flag 
61. Qual ( Dbl_Soln_Qual ) - Solution quality flag 
62. m_HIP ( Dbl_Ref_ID ) - Component identifiers 
63. theta ( Dbl_Theta ) - Position angle between components 
64. rho ( Dbl_Rho ) - Angular separation of components 
65. e_rho ( Rho_Error ) - Standard error of rho 
66. dHp ( Diff_Hip_Mag ) - Magnitude difference of components 
67. e_dHp ( dHip_Mag_Error ) - Standard error in dHp 
68. Survey ( Survey_Star ) - Flag indicating a Survey Star 
69. Chart ( ID_Chart ) - Identification Chart 
70. Notes ( Notes ) - Existence of notes 
71. HD ( HD_Id ) - HD number (III 135) 
72. BD ( BD_Id ) - Bonner DM (I 119), (I 122) 
73. CoD ( CoD_Id ) - Cordoba Durchmusterung (DM) (I 114) 
74. CPD ( CPD_Id ) - Cape Photographic DM (I 108) 
75. (V-I)red ( VI_Color_Reduct ) - VI used for reductions 
76. SpType ( Spect_Type ) - Spectral type 
77. r_SpType ( Spect_Type_Source ) - Source of spectral type 
*/

interface Star {
	_id: string;
	Catalog: string;
	HIP: number;
	Proxy: string;
	RAhms: string;
	DEdms: string;
	Vmag: number;
	VarFlag: string;
	r_Vmag: string;
	RAdeg: number;
	DEdeg: number;
	AstroRef: string;
	Plx: number;
	pmRA: number;
	pmDE: number;
	e_RAdeg: number;
	e_DEdeg: number;
	e_Plx: number;
	e_pmRA: number;
	e_pmDE: number;
	'DE:RA': number;
	'Plx:RA': number;
	'Plx:DE': number;
	'pmRA:RA': number;
	'pmRA:DE': number;
	'pmRA:Plx': number;
	'pmDE:RA': number;
	'pmDE:DE': number;
	'pmDE:Plx': number;
	'pmDE:pmRA': number;
	F1: number;
	F2: number;
	'---': number;
	BTmag: number;
	e_BTmag: number;
	VTmag: number;
	e_VTmag: number;
	m_BTmag: string;
	'B-V': number;
	'e_B-V': number;
	'r_B-V': string;
	'V-I': number;
	'e_V-I': number;
	'r_V-I': string;
	CombMag: string;
	Hpmag: number;
	e_Hpmag: number;
	Hpscat: number;
	o_Hpmag: number;
	m_Hpmag: string;
	Hpmax: number;
	HPmin: number;
	Period: number | null;
	HvarType: string;
	moreVar: string | null;
	morePhoto: string | null;
	CCDM: string;
	n_CCDM: string;
	Nsys: number | null;
	Ncomp: number | null;
	MultFlag: string;
	Source: string;
	Qual: string;
	m_HIP: string;
	theta: number | null;
	rho: number | null;
	e_rho: number | null;
	dHp: number | null;
	e_dHp: number | null;
	Survey: string;
	Chart: string;
	Notes: string;
	HD: number | null;
	BD: string;
	CoD: string;
	CPD: string;
	'(V-I)red': number | null;
	SpType: string;
	r_SpType: string;
}

const bv2rgb = (bv: number): string => {
	bv = Math.max(-0.4, Math.min(2, bv));
	let t;
	return `#${[
		bv < 0
			? ((t = (bv + 0.4) / 0.4), 0.61 + 0.11 * t + 0.1 * t * t)
			: bv < 0.4
			? ((t = bv / 0.4), 0.83 + 0.17 * t)
			: 1,
		bv < 0
			? ((t = (bv + 0.4) / 0.4), 0.7 + 0.07 * t + 0.1 * t * t)
			: bv < 0.4
			? ((t = bv / 0.4), 0.87 + 0.11 * t)
			: bv < 1.6
			? ((t = (bv - 0.4) / 1.2), 0.98 - 0.16 * t)
			: ((t = (bv - 1.6) / 0.4), 0.82 - 0.5 * t * t),
		bv < 0.4
			? 1
			: bv < 1.5
			? ((t = (bv - 0.4) / 1.1), 1 - 0.47 * t + 0.1 * t * t)
			: bv < 1.94
			? ((t = (bv - 1.5) / 0.44), 0.63 - 0.6 * t * t)
			: 0
	]
		.map((t) =>
			Math.round(t * 255)
				.toString(16)
				.padStart(2, '0')
		)
		.join('')}`;
};

const temperature = (color: number): number => {
	return 4600 * (1 / (0.92 * color + 1.7) + 1 / (0.92 * color + 0.62));
};

export { type StarHR, type Star, bv2rgb, temperature };
