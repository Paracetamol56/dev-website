import { PDFDocument, PDFFont, PDFPage, rgb } from 'pdf-lib';

export type PrintingMarksOptions = {
	unit: 'mm' | 'in';
	cropMarks: boolean;
	bleeds: {
		top: number;
		right: number;
		bottom: number;
		left: number;
	} | null;
	registrationMarks: boolean;
	colorBars: boolean;
	pageInfo: {
		pageNumbering: boolean;
		pageInfoText: string;
	} | null;
};

// PDF processing
export async function generatePDF(
	uploadedFile: File,
	options: PrintingMarksOptions
): Promise<Blob> {
	// Load the PDF and add margins
	return new Promise((resolve, reject) => {
		const fileReader = new FileReader();
		fileReader.onload = async function () {
			const existingPdfBytes = new Uint8Array(this.result as ArrayBuffer);
			const pdfDoc = await PDFDocument.load(existingPdfBytes);
			const newPdf = await PDFDocument.create();
			const margin = 34;

			// Process each page
			for (let i = 0; i < pdfDoc.getPageCount(); i++) {
				const page = pdfDoc.getPage(i);

				const newPage: PDFPage = newPdf.addPage([
					page.getWidth() + 2 * margin,
					page.getHeight() + 2 * margin
				]);

				const embeddedPage = await newPdf.embedPage(page);

				newPage.drawPage(embeddedPage, {
					x: margin,
					y: margin,
					xScale: 1,
					yScale: 1
				});

				if (options.cropMarks) {
					drawCropMarks(
						newPage,
						margin,
						options.unit,
						options.bleeds || { top: 0, right: 0, bottom: 0, left: 0 }
					);
				}

				if (options.registrationMarks) {
					drawRegistrationMarks(newPage, margin);
				}

				if (options.colorBars) {
					await drawColorBars(newPdf, newPage, margin); // Position above black tint bar
				}

				if (options.pageInfo) {
					await drawPageInfo(newPdf, margin, newPage, i, pdfDoc, options.pageInfo);
				}
			}

			// Save the PDF
			const pdfBytes = (await newPdf.save()) as BlobPart;
			const blob = new Blob([pdfBytes], { type: 'application/pdf' });
			resolve(blob);
		};
		fileReader.onerror = () => {
			reject(new Error('Failed to read file'));
		};
		fileReader.readAsArrayBuffer(uploadedFile);
	});
}

async function drawPageInfo(
	newPdf: PDFDocument,
	margin: number,
	page: PDFPage,
	i: number,
	pdfDoc: PDFDocument,
	options: {
		pageNumbering: boolean;
		pageInfoText: string;
	}
) {
	const font = await newPdf.embedFont('Helvetica');
	const fontSize = 9;

	// Position page info text in the bottom margin
	const pageInfoY = margin - 20;

	// Left page info (aligned with left color bar)
	const leftPageInfoX = margin + 40;

	// Right page info (aligned with right color bar)
	const rightPageInfoX = page.getWidth() - (margin + 40 + 45);

	// Add page numbering if enabled
	if (options.pageNumbering) {
		const pageNumberText = `Page ${i + 1} of ${pdfDoc.getPageCount()}`;
		const textWidth = font.widthOfTextAtSize(pageNumberText, fontSize);

		// Trim text if it's too long to fit in the available space
		let trimmedText = pageNumberText;
		const maxWidth = 14 * 17 - 10;
		if (textWidth > maxWidth) {
			const charsPerPt = pageNumberText.length / textWidth;
			const maxChars = Math.floor(maxWidth * charsPerPt);
			trimmedText = pageNumberText.substring(0, maxChars - 3) + '...';
		}

		// Draw page numbering on the left
		page.drawText(trimmedText, {
			x: rightPageInfoX,
			y: pageInfoY,
			size: fontSize,
			font,
			color: rgb(0, 0, 0)
		});
	}

	// Add custom page info text if enabled
	if (options.pageInfoText) {
		const textWidth = font.widthOfTextAtSize(options.pageInfoText, fontSize);

		// Trim text if it's too long to fit in the available space
		let trimmedText = options.pageInfoText;
		const maxWidth = 14 * 17 - 10;
		if (textWidth > maxWidth) {
			const charsPerPt = options.pageInfoText.length / textWidth;
			const maxChars = Math.floor(maxWidth * charsPerPt);
			trimmedText = options.pageInfoText.substring(0, maxChars - 3) + '...';
		}

		// Draw custom text on the right
		page.drawText(trimmedText, {
			x: leftPageInfoX,
			y: pageInfoY,
			size: fontSize,
			font,
			color: rgb(0, 0, 0)
		});
	}
}

function drawBlackTintBar(page: PDFPage, x: number, y: number, font: PDFFont, fontSize: number) {
	const cellWidth = 14;
	const cellHeight = 14;

	const tints = [
		{ label: '0%', opacity: 0 },
		{ label: '10%', opacity: 0.1 },
		{ label: '20%', opacity: 0.2 },
		{ label: '40%', opacity: 0.4 },
		{ label: '50%', opacity: 0.5 },
		{ label: '60%', opacity: 0.6 },
		{ label: '70%', opacity: 0.7 },
		{ label: '80%', opacity: 0.8 },
		{ label: '90%', opacity: 0.9 },
		{ label: '100%', opacity: 1.0 }
	];

	tints.forEach((tint, i) => {
		const cellX = x + i * cellWidth;
		const rectY = y; // bottom of rect in PDF space
		const textY = rectY + cellHeight + 2; // text above the rect
		const textX = cellX + cellWidth / 2;

		// Label
		const textWidth = font.widthOfTextAtSize(tint.label, fontSize);
		page.drawText(tint.label, {
			x: textX - textWidth / 2,
			y: textY,
			size: fontSize,
			font,
			color: rgb(0, 0, 0)
		});

		// Rectangle
		page.drawRectangle({
			x: cellX,
			y: rectY,
			width: cellWidth,
			height: cellHeight,
			color: rgb(0, 0, 0),
			opacity: tint.opacity
		});
	});
}

function drawColorMixBar(page: PDFPage, x: number, y: number, font: PDFFont, fontSize: number) {
	const cellWidth = 14;
	const cellHeight = 14;

	// Helper to convert hex to rgb (0-1 range) for pdf-lib
	const hexToRgb = (hex: string) => {
		const bigint = parseInt(hex.replace('#', ''), 16);
		const r = ((bigint >> 16) & 255) / 255;
		const g = ((bigint >> 8) & 255) / 255;
		const b = (bigint & 255) / 255;
		return rgb(r, g, b);
	};

	const swatches = [
		{ label: 'CMY', color: '#231F20' },
		{ label: 'K', color: '#000000' },
		{ label: 'C40', color: '#00AEEF', opacity: 0.4 },
		{ label: 'C60', color: '#00AEEF', opacity: 0.6 },
		{ label: 'C80', color: '#00AEEF', opacity: 0.8 },
		{ label: 'C', color: '#00AEEF' },
		{ label: 'M40', color: '#EC008C', opacity: 0.4 },
		{ label: 'M60', color: '#EC008C', opacity: 0.6 },
		{ label: 'M80', color: '#EC008C', opacity: 0.8 },
		{ label: 'M', color: '#EC008C' },
		{ label: 'Y40', color: '#FFF200', opacity: 0.4 },
		{ label: 'Y60', color: '#FFF200', opacity: 0.6 },
		{ label: 'Y80', color: '#FFF200', opacity: 0.8 },
		{ label: 'Y', color: '#FFF200' },
		{ label: 'CM', color: '#2E3192' },
		{ label: 'CY', color: '#00A651' },
		{ label: 'MY', color: '#ED1C24' }
	];

	swatches.forEach((swatch, i) => {
		const cellX = x + i * cellWidth;
		const rectY = y;
		const textY = rectY + cellHeight + 2;
		const textX = cellX + cellWidth / 2;

		const textWidth = font.widthOfTextAtSize(swatch.label, fontSize);
		page.drawText(swatch.label, {
			x: textX - textWidth / 2,
			y: textY,
			size: fontSize,
			font,
			color: rgb(0, 0, 0)
		});

		page.drawRectangle({
			x: cellX,
			y: rectY,
			width: cellWidth,
			height: cellHeight,
			color: hexToRgb(swatch.color),
			opacity: swatch.opacity
		});
	});
}

async function drawColorBars(newPdf: PDFDocument, page: PDFPage, margin: number) {
	const font: PDFFont = await newPdf.embedFont('Helvetica');
	const fontSize = 6;

	const regMarkSize = 14;
	const regMarkOffset = regMarkSize / 2;

	const colorBarY = page.getHeight() - margin + regMarkOffset;

	// Left color bar (aligned with left registration mark)
	const leftColorBarX = margin + 40;
	drawBlackTintBar(page, leftColorBarX, colorBarY, font, fontSize);

	// Right color bar (aligned with right registration mark)
	const rightColorBarX = page.getWidth() - (margin + 40 + 14 * 17);
	drawColorMixBar(page, rightColorBarX, colorBarY, font, fontSize);
}

function drawRegistrationMarks(page: PDFPage, margin: number): void {
	const regMarkSize = 35;
	const regMarkOffset = regMarkSize / 2;
	const offset = 15;
	const centerX = regMarkSize / 2;
	const outerRadius = 8;
	const innerRadius = 5;
	const lineLength = 25;

	// Function to draw a registration mark at a specific position
	const drawRegistrationMark = (x: number, y: number) => {
		const cx = x + centerX;
		const cy = y + centerX;

		// Outer circle (r=18 equivalent -> outerRadius)
		page.drawCircle({
			x: cx,
			y: cy,
			size: outerRadius,
			borderColor: rgb(0, 0, 0),
			borderWidth: 0.5
		});

		// Full-length vertical line (black, spans beyond outer circle)
		page.drawLine({
			start: { x: cx, y: cy - lineLength / 2 },
			end: { x: cx, y: cy + lineLength / 2 },
			thickness: 0.5,
			color: rgb(0, 0, 0)
		});

		// Full-length horizontal line (black, spans beyond outer circle)
		page.drawLine({
			start: { x: cx - lineLength / 2, y: cy },
			end: { x: cx + lineLength / 2, y: cy },
			thickness: 0.5,
			color: rgb(0, 0, 0)
		});

		// Filled black inner circle (covers crosshair where they overlap)
		page.drawCircle({
			x: cx,
			y: cy,
			size: innerRadius,
			borderColor: rgb(0, 0, 0),
			borderWidth: 0.5,
			color: rgb(0, 0, 0)
		});

		// Short white vertical line segment (only within inner circle)
		page.drawLine({
			start: { x: cx, y: cy - innerRadius },
			end: { x: cx, y: cy + innerRadius },
			thickness: 0.5,
			color: rgb(1, 1, 1)
		});

		// Short white horizontal line segment (only within inner circle)
		page.drawLine({
			start: { x: cx - innerRadius, y: cy },
			end: { x: cx + innerRadius, y: cy },
			thickness: 0.5,
			color: rgb(1, 1, 1)
		});
	};

	// Top center (offset into margin)
	drawRegistrationMark(
		page.getWidth() / 2 - regMarkOffset,
		page.getHeight() - margin + offset - regMarkOffset
	);

	// Bottom center (offset into margin)
	drawRegistrationMark(page.getWidth() / 2 - regMarkOffset, margin - offset - regMarkOffset);

	// Left center (offset into margin)
	drawRegistrationMark(margin - offset - regMarkOffset, page.getHeight() / 2 - regMarkOffset);

	// Right center (offset into margin)
	drawRegistrationMark(
		page.getWidth() - margin + offset - regMarkOffset,
		page.getHeight() / 2 - regMarkOffset
	);
}

function drawCropMarks(
	page: PDFPage,
	margin: number,
	unit: 'mm' | 'in',
	bleeds: {
		top: number;
		right: number;
		bottom: number;
		left: number;
	}
): void {
	const unitMultiplier = unit === 'mm' ? 2.83465 : 72;
	const bleedTopPts = bleeds.top * unitMultiplier;
	const bleedRightPts = bleeds.right * unitMultiplier;
	const bleedBottomPts = bleeds.bottom * unitMultiplier;
	const bleedLeftPts = bleeds.left * unitMultiplier;

	// Draw crop marks at each corner
	// Top-left corner
	page.drawLine({
		start: { x: margin - 25, y: page.getHeight() - margin - bleedTopPts },
		end: { x: margin - 5, y: page.getHeight() - margin - bleedTopPts },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});
	page.drawLine({
		start: { x: margin + bleedLeftPts, y: page.getHeight() - margin + 25 },
		end: { x: margin + bleedLeftPts, y: page.getHeight() - margin + 5 },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});

	// Top-right corner
	page.drawLine({
		start: {
			x: page.getWidth() - margin + 5,
			y: page.getHeight() - margin - bleedTopPts
		},
		end: {
			x: page.getWidth() - margin + 25,
			y: page.getHeight() - margin - bleedTopPts
		},
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});
	page.drawLine({
		start: {
			x: page.getWidth() - margin - bleedRightPts,
			y: page.getHeight() - margin + 25
		},
		end: {
			x: page.getWidth() - margin - bleedRightPts,
			y: page.getHeight() - margin + 5
		},
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});

	// Bottom-left corner
	page.drawLine({
		start: { x: margin - 25, y: margin + bleedBottomPts },
		end: { x: margin - 5, y: margin + bleedBottomPts },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});
	page.drawLine({
		start: { x: margin + bleedLeftPts, y: margin - 25 },
		end: { x: margin + bleedLeftPts, y: margin - 5 },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});

	// Bottom-right corner
	page.drawLine({
		start: { x: page.getWidth() - margin + 5, y: margin + bleedBottomPts },
		end: { x: page.getWidth() - margin + 25, y: margin + bleedBottomPts },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});
	page.drawLine({
		start: { x: page.getWidth() - margin - bleedRightPts, y: margin - 25 },
		end: { x: page.getWidth() - margin - bleedRightPts, y: margin - 5 },
		thickness: 0.5,
		color: rgb(0, 0, 0)
	});
}
