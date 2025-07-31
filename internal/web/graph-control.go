package web

// Can not include javascript file into a resulting binary template after go compile,
// so use go code for js.
// TODO: think about better solution. Is it posible to improve?

const GraphControlJS = `
// TODO: doesn't work since we remove attributes later (width, height)
// think about good scroll/resize/focus mechanics
const svg = document.querySelector('#svg');

const btnZoomIn = document.querySelector('#zoom-in');
const btnZoomOut = document.querySelector('#zoom-out');

btnZoomIn.addEventListener('click', () => {
    resize(1.1);
});

btnZoomOut.addEventListener('click', () => {
    resize(0.9);
});

function resize(scale) {
    let svgWidth = parseInt(svg.getAttribute('width'));
    svg.setAttribute('width', ` + "`${(svgWidth * scale)}`" + `);
    let svgHeight = parseInt(svg.getAttribute('height'));
    svg.setAttribute('height', ` + "`${(svgHeight * scale)}`" + `);
}

// For wider picture angle
// TODO: maybe remove
svgCnvs = document.getElementById('svg')
svgCnvs.removeAttribute('width')
svgCnvs.removeAttribute('height')

// This class allows to zoom in to graph node on tree element click.
// it works almost good, but there's a lot of things to do:
// TODO:
// - add opportunity to scroll left and right after focus
// - not really good at big projects like kubernetes,
//  sometimes need to click twice to zoom properly.
//  Probably caused by zoom reset before zoom apply. Need to investigate.
class SVGZoomController {
	constructor(svgElement, containerElement, options = {}) {
		this.svg = svgElement;
		this.container = containerElement;
		this.initialViewBox = this.parseViewBox(this.svg.getAttribute('viewBox'));
		this.currentViewBox = {...this.initialViewBox};
		this.zoomFactor = options.zoomFactor || 1.5;
		this.maxZoomLevel = options.maxZoomLevel || 5;
		this.zoomLevel = 0;
		this.animationDuration = options.animationDuration || 300;
		
		this.init();
	}
	
	parseViewBox(viewBoxStr) {
		const parts = viewBoxStr.split(' ').map(Number);
		return {
			x: parts[0],
			y: parts[1],
			width: parts[2],
			height: parts[3]
		};
	}
	
	init() {
		this.svg.addEventListener('click', (e) => {
			if (e.target.tagName !== 'svg') {
				this.zoomToElement(e.target);
			}
		});
		
		document.getElementById('resetZoom')?.addEventListener('click', () => this.resetZoom());
	}
	
	zoomToElement(element) {
		if (this.zoomLevel >= this.maxZoomLevel) return;

		// Otherwise second zoom will be weird.
		this.resetZoom();
		
		const bbox = element.getBBox();
		// GraphViz nodes have negative Y. But if we sustract modulus Y from the height,
		// we obtain desired positive value.
        if (bbox.y < 0) {
            bbox.y = this.currentViewBox.height + bbox.y
        }

		const centerX = bbox.x + bbox.width / 2;
		const centerY = bbox.y + bbox.height / 2;
		
		// Calculate new viewBox
		const newWidth = this.currentViewBox.width / this.zoomFactor;
		const newHeight = this.currentViewBox.height / this.zoomFactor;
		
		const newViewBox = {
			x: centerX - newWidth / 2,
			y: centerY - newHeight / 2,
			width: newWidth,
			height: newHeight
		};
		
		// Animating viewBox change
		this.animateViewBox(this.currentViewBox, newViewBox);
		
		this.currentViewBox = newViewBox;
		this.zoomLevel++;
		
		// After animation fix scroll
		setTimeout(() => {
			this.adjustScrollPosition(bbox);
		}, this.animationDuration);
	}
	
	animateViewBox(from, to) {
		const startTime = performance.now();
		const duration = this.animationDuration;
		
		const animate = (currentTime) => {
			const elapsed = currentTime - startTime;
			const progress = Math.min(elapsed / duration, 1);
			
			const interpolated = {
				x: from.x + (to.x - from.x) * progress,
				y: from.y + (to.y - from.y) * progress,
				width: from.width + (to.width - from.width) * progress,
				height: from.height + (to.height - from.height) * progress
			};
			
			this.svg.setAttribute('viewBox',
				` + "`${interpolated.x} ${interpolated.y} ${interpolated.width} ${interpolated.height}`" + `);
			
			if (progress < 1) {
				requestAnimationFrame(animate);
			}
		};
		
		requestAnimationFrame(animate);
	}
	
	adjustScrollPosition(bbox) {
		// Obtain element position relative to container
		const svgRect = this.svg.getBoundingClientRect();
		const containerRect = this.container.getBoundingClientRect();
		
		// Calculate center of the element
		const elementCenterX = bbox.x + bbox.width / 2;
		const elementCenterY = bbox.y + bbox.height / 2;
		
		// Convert SVG coordinates to viewport coordinates
		const viewBox = this.currentViewBox;
		const scaleX = svgRect.width / viewBox.width;
		const scaleY = svgRect.height / viewBox.height;
		
		const viewportX = (elementCenterX - viewBox.x) * scaleX;
		const viewportY = (elementCenterY - viewBox.y) * scaleY;
		
		// Calculate new scroll position
		const targetScrollLeft = viewportX + svgRect.left - containerRect.left - containerRect.width / 2;
		const targetScrollTop = viewportY + svgRect.top - containerRect.top - containerRect.height / 2;
		
		// Smoothly scroll to the desired position
		this.container.scrollTo({
			left: targetScrollLeft,
			top: targetScrollTop,
			behavior: 'smooth'
		});
	}
	
	resetZoom() {
		this.animateViewBox(this.currentViewBox, this.initialViewBox);
		this.currentViewBox = {...this.initialViewBox};
		this.zoomLevel = 0;
		
		// Reset scroll to the begining
		setTimeout(() => {
			this.container.scrollTo({
				top: 0,
				left: 0,
				behavior: 'smooth'
			});
		}, this.animationDuration);
	}

}


// Init after DOM loaded
document.addEventListener('DOMContentLoaded', () => {
	// Init zoomer
	const svg = document.getElementById('svg');
	const container = document.getElementById('container');
	const zoomController = new SVGZoomController(svg, container, {
		zoomFactor: 2,
		maxZoomLevel: 8,
		animationDuration: 300
	});

	// Gather all graph nodes to match them with tree nodes.
	// This is dirty hacks because I used 'tree' util to generate inital html.
	// When there will be native go tree generation, this code will be or refactored.
	var graphNodes = document.querySelectorAll('g[class="node"]');
	var nodeTitles = {};
	for (var i = 0; i < graphNodes.length; i++) {
		title = graphNodes[i].childNodes[1].innerHTML.replace(/"/g, '');
		nodeID = graphNodes[i].id;
		nodeTitles[title] = nodeID;
	}

	// Create map with [tree node id] graph node id
	var anchors = document.getElementsByTagName('a');
	var treeGraphMap = {};
	for (var i = 0; i < anchors.length; i++) {
		const anchor = anchors[i];
		if (anchor.className == "DIR") {
			relativePackagePath = anchor.getAttribute('href').slice(1, -1);
			anchor.removeAttribute('href')
			anchor.setAttribute('id', relativePackagePath)

			found = false;
			for (const packagePath in nodeTitles) {
				if (packagePath.endsWith(relativePackagePath)) {
					const nodeID = nodeTitles[packagePath];
					treeGraphMap[anchor.id] = nodeID

					found = true;
					
					break;
				}
			}
		}
	}

	// Set element tree click action - zoom and highlight
	for (var i = 0; i < anchors.length; i++) {
		const anchor = anchors[i];
		anchor.onclick = function() {
			graphNodeID = treeGraphMap[anchor.id];

			graphNode = document.getElementById(graphNodeID);
			if (!graphNode) {
				// For a while print alert. Later create differen color
				// for simple directories (not packages).
				alert(` + "`${anchor.id}`" + ` + " is not a package");
				return;
			}
			zoomController.zoomToElement(graphNode);

			polygon = graphNode.getElementsByTagName("polygon")[0];
			originFill = polygon.getAttribute("fill");
			polygon.setAttribute("fill", "red");

			setTimeout(() => {
				polygon.setAttribute("fill", originFill);
			  }, "1000");
		};
	}
});

`
