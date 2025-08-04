// For initial wide angle
svgCnvs = document.getElementById('svg')
svgCnvs.removeAttribute('width')
svgCnvs.removeAttribute('height')

// This class allows to zoom in to graph node on tree element click.
// it works almost good, but there's a lot of things to do:
class SVGViewController {
	constructor(svgElement, containerElement, options = {}) {
		this.svg = svgElement;
		this.container = containerElement;
		this.initialViewBox = this.parseViewBox(this.svg.getAttribute('viewBox'));
		this.currentViewBox = {...this.initialViewBox};
		this.zoomFactor = options.zoomFactor || 1.5;
        this.zoomElementFactor = options.zoomElementFactor || options.zoomFactor || 1.5;
        this.slowZoomFactor = 1.05;
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
        // Handle keyboard: + and -
        document.addEventListener('keydown', (e) => {
            // + and = are on the same key
            if (e.key == "+" || e.key == "=") {
                this.zoomIn()
            }
            if (e.key == "-") {
                this.zoomOut()
            }
        });

        // Controls
        document.getElementById('resetZoom')?.addEventListener('click', () => this.resetZoom());
        document.getElementById('zoomIn')?.addEventListener('click', () => this.zoomIn());
        document.getElementById('zoomOut')?.addEventListener('click', () => this.zoomOut());
                
        // Mouse pan
        this.container.addEventListener('mousedown', this.startPan.bind(this));
        document.addEventListener('mousemove', this.doPan.bind(this));
        document.addEventListener('mouseup', this.endPan.bind(this));
        document.addEventListener('mouseleave', this.endPan.bind(this));
                
        // Mouse wheel pan
        this.container.addEventListener('wheel', this.handleWheel.bind(this), { passive: false });

        // Wheel zoom
        this.container.addEventListener('DOMMouseScroll', this.handleScroll, false); // for Firefox
        this.container.addEventListener('mousewheel', this.handleScroll, false); // for everyone else
                
        // Cursor change on pan
        this.svg.style.cursor = 'grab';
    }

    startPan(e) {
        if (e.button !== 0) return; // Left mouse button only
                
        this.isPanning = true;
        this.startPanPoint = { x: e.clientX, y: e.clientY };
        this.panStartViewBox = { ...this.currentViewBox };
        this.svg.style.cursor = 'grabbing';
        e.preventDefault();
    }

    doPan(e) {
        if (!this.isPanning) return;
        
        const dx = (e.clientX - this.startPanPoint.x) * (this.currentViewBox.width / this.container.clientWidth);
        const dy = (e.clientY - this.startPanPoint.y) * (this.currentViewBox.height / this.container.clientHeight);
        
        this.currentViewBox.x = this.panStartViewBox.x - dx;
        this.currentViewBox.y = this.panStartViewBox.y - dy;
        
        this.updateViewBox();
        e.preventDefault();
    }
    
    endPan(e) {
        if (!this.isPanning) return;
        
        this.isPanning = false;
        this.svg.style.cursor = 'grab';
        e.preventDefault();
    }

    handleScroll(e) {
        var direction = (e.detail<0 || e.wheelDelta>0) ? 1 : -1;
        const zoomPoint = this.getSVGPoint(e.clientX, e.clientY);

        if (direction > 0) {
            this.zoomToPoint(zoomPoint.x, zoomPoint.y, 1 / this.slowZoomFactor);
        } else if (direction < 0) {
            this.zoomToPoint(zoomPoint.x, zoomPoint.y, this.slowZoomFactor);
        }

        e.preventDefault();
    };
    
    handleWheel(e) {
        if (e.ctrlKey) {
            // Scale with wheel on pressed Ctrl
            this.handleScroll(e)
        } else {
            // Simple wheel pan
            const panSpeed = 0.5;
            this.currentViewBox.x += e.deltaX * panSpeed * (this.currentViewBox.width / this.container.clientWidth);
            this.currentViewBox.y += e.deltaY * panSpeed * (this.currentViewBox.height / this.container.clientHeight);
            this.updateViewBox();
            e.preventDefault();
        }
    }

    getSVGPoint(clientX, clientY) {
        const rect = this.svg.getBoundingClientRect();
        const x = clientX - rect.left;
        const y = clientY - rect.top;
        
        return {
            x: this.currentViewBox.x + (x / rect.width) * this.currentViewBox.width,
            y: this.currentViewBox.y + (y / rect.height) * this.currentViewBox.height
        };
    }

    zoomToPoint(x, y, factor) {
        const newWidth = this.currentViewBox.width * factor;
        const newHeight = this.currentViewBox.height * factor;
        
        const newX = x - (x - this.currentViewBox.x) * factor;
        const newY = y - (y - this.currentViewBox.y) * factor;
        
        this.currentViewBox = {
            x: newX,
            y: newY,
            width: newWidth,
            height: newHeight
        };
        
        this.zoomLevel += factor > 1 ? 1 : -1;
        this.updateViewBox();
    }
    
    zoomOut() {        
        const containerRect = this.container.getBoundingClientRect();
        const centerX = containerRect.width / 2;
        const centerY = containerRect.height / 2;
        const center = this.getSVGPoint(containerRect.left + centerX, containerRect.top + centerY);
        
        this.zoomToPoint(center.x, center.y, this.zoomFactor);
    }
    
    zoomIn() {        
        const containerRect = this.container.getBoundingClientRect();
        const centerX = containerRect.width / 2;
        const centerY = containerRect.height / 2;
        const center = this.getSVGPoint(containerRect.left + centerX, containerRect.top + centerY);
        
        this.zoomToPoint(center.x, center.y, 1 / this.zoomFactor);
    }
    
    zoomToElement(element) {
        // Otherwise next zoom will be weird.
		this.resetZoom();
        
        let bbox = element.getBBox();
        // GraphViz nodes have negative Y. But if we sustract modulus Y from the height,
		// we obtain desired positive value.
        if (bbox.y < 0) {
            bbox.y = this.currentViewBox.height + bbox.y
        }

        const centerX = bbox.x + bbox.width / 2;
        const centerY = bbox.y + bbox.height / 2;
        
        this.zoomToPoint(centerX, centerY, 1 / this.zoomElementFactor);
        this.adjustScrollPosition(bbox);
    }
    
    adjustScrollPosition(bbox) {
        // Get element position relative to container
        const svgRect = this.svg.getBoundingClientRect();
        const containerRect = this.container.getBoundingClientRect();

        // Negative Y causes non-idempotent scroll to element.
        if (svgRect.y < 0) {
            svgRect.y = svgRect.height + svgRect.y
        }
        
        // Calculate element center
        const elementCenterX = bbox.x + bbox.width / 2;
        const elementCenterY = bbox.y + bbox.height / 2;
        
        // Convert SVG coordinates to viewport coordinates
        const viewBox = this.currentViewBox;
        const scaleX = svgRect.width / viewBox.width;
        const scaleY = svgRect.height / viewBox.height;
        
        const viewportX = (elementCenterX - viewBox.x) * scaleX;
        const viewportY = (elementCenterY - viewBox.y) * scaleY;
        
        // Calculate new scroll bar positions
        const targetScrollLeft = viewportX + svgRect.left - containerRect.left - containerRect.width / 2;
        const targetScrollTop = viewportY + svgRect.top - containerRect.top - containerRect.height / 2;
        
        // Smooth scroll to desired position
        this.container.scrollTo({
            left: targetScrollLeft,
            top: targetScrollTop,
            behavior: 'smooth'
        });
    }
    
    resetZoom() {
        this.currentViewBox = {...this.initialViewBox};
        this.updateViewBox();
        this.zoomLevel = 0;
        
        // Reset scroll to begining
        this.container.scrollTo({
            top: 0,
            left: 0,
            behavior: 'smooth'
        });
    }
    
    updateViewBox() {
        const {x, y, width, height} = this.currentViewBox;
        this.svg.setAttribute('viewBox', `${x} ${y} ${width} ${height}`);
    }
}

// Need to zoom to the same scale for any svg size.
function calculateZoomFactor(svg) {
    const svgWidth = svg.getBBox().width;
    // Found 0.0008 empirically
    let factor = svgWidth * 0.0008
    if (factor < 1) {
        factor = 1
    }
    return factor
}

// Init after DOM loaded
document.addEventListener('DOMContentLoaded', () => {
	// Init zoomer
	const svg = document.getElementById('svg');
	const container = document.getElementById('svgContainer');
	const zoomController = new SVGViewController(svg, container, {
		zoomFactor: 1.5,
        zoomElementFactor: calculateZoomFactor(svg),
		animationDuration: 300
	});

    // TODO: refactor code below.

	// Gather all graph nodes to match them with tree nodes.
	var graphNodes = document.querySelectorAll('g[class="node"]');
	var nodeTitles = {};
	for (var i = 0; i < graphNodes.length; i++) {
		title = graphNodes[i].childNodes[1].innerHTML.replace(/"/g, '');
		nodeID = graphNodes[i].id;
		nodeTitles[title] = nodeID;
	}

	// Create map with [tree node id] graph node id
    var anchors = document.getElementsByClassName("tree-entry")
	var treeGraphMap = {};
	for (var i = 0; i < anchors.length; i++) {
		const anchor = anchors[i];
        const relativePackagePath = anchor.id
        for (const packagePath in nodeTitles) {
            if (packagePath.endsWith(relativePackagePath)) {
                const nodeID = nodeTitles[packagePath];
                treeGraphMap[relativePackagePath] = nodeID

                break;
            }
        }
	}

	// Set element tree click action - zoom and highlight
    var anchors = document.getElementsByClassName("tree-entry")
	for (var i = 0; i < anchors.length; i++) {
		const anchor = anchors[i];
		anchor.onclick = function() {
			const graphNodeID = treeGraphMap[anchor.id];

			const graphNode = document.getElementById(graphNodeID);
			if (!graphNode) {
				return;
			}
			zoomController.zoomToElement(graphNode);

			const polygon = graphNode.getElementsByTagName("polygon")[0];
			polygon.setAttribute("fill", "#FACDEE");

			const text = graphNode.getElementsByTagName("text")[0];
			text.setAttribute("font-size", 14);

			setTimeout(() => {
				polygon.setAttribute("fill", "none");
                text.setAttribute("font-size", 10);
			  }, "3000");
		};
	}

    // TODO: remove this after native graph generate
    const nodes = document.getElementsByClassName("node")
    for (var i = 0; i < nodes.length; i++) {
        const node = nodes[i]
        const link = node.getElementsByTagName("a");

        // Remove click action
        link[0].removeAttribute("xlink:href")
    }
});
