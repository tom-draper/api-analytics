function initSVG() {
  let svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("width", 1000);
  svg.setAttribute("height", 1000);
  svg.id = "mySVG";
  return svg;
}

function randn_bm(min, max, skew) {
  let u = 0,
    v = 0;
  while (u === 0) u = Math.random(); //Converting [0,1) to (0,1)
  while (v === 0) v = Math.random();
  let num = Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);

  num = num / 10.0 + 0.5; // Translate to 0 -> 1
  if (num > 1 || num < 0)
    num = randn_bm(min, max, skew); // resample between 0 and 1 if out of range
  else {
    num = Math.pow(num, skew); // Skew
    num *= max - min; // Stretch to fill range
    num += min; // offset to min
  }
  return num;
}

function gaussian(mean, stdev) {
  var y2;
  var use_last = false;
  var y1;
  if (use_last) {
    y1 = y2;
    use_last = false;
  } else {
    var x1, x2, w;
    do {
      x1 = 2.0 * Math.random() - 1.0;
      x2 = 2.0 * Math.random() - 1.0;
      w = x1 * x1 + x2 * x2;
    } while (w >= 1.0);
    w = Math.sqrt((-2.0 * Math.log(w)) / w);
    y1 = x1 * w;
    y2 = x2 * w;
    use_last = true;
  }

  var retval = mean + stdev * y1;
  if (retval > 0) return retval;
  return -retval;
}

function setCirclePosition(circle, mean, std) {
  let x = gaussian(mean, std);
  let y = gaussian(mean, std);
  circle.setAttributeNS(null, "cx", x);
  circle.setAttributeNS(null, "cy", y);
}

function randomInt(low, high) {
  return Math.floor(Math.random() * high) + low;
}

function setCircleSize(circle) {
  let r = randomInt(1, 5);
  circle.setAttributeNS(null, "r", r);
}

function distance(x1, y1, x2, y2) {
  return Math.sqrt((Math.pow(x1-x2, 2))+(Math.pow(y1-y2, 2)))
}

function randomChoice(p) {
    let rnd = p.reduce( (a, b) => a + b ) * Math.random();
    return p.findIndex( a => (rnd -= a) < 0 );
}

function randomChoices(p, count) {
    return Array.from(Array(count), randomChoice.bind(null, p));
}

let color = [
  '#1c1c1c',
  '#05140d',
  '#0f3d28',
  '#196643',
  '#248f5e',
  '#2eb879',
  '#47d193',
  '#47d193',
  '#99e6c3',
  '#99e6c3',
  '#99e6c3',


  // '#232222',
  // '#242624',
  // 'green',
  // '#3fcf8e',
  // 'lightgreen',
  // 'lime',
  // 'yellow',
  // 'red',
]

function setCircleColor(circle) {
  let d = distance(circle.cx.baseVal.value, circle.cy.baseVal.value, 500, 500);
  let p = Array(color.length).fill(0)
  if (d < 80) {
    p[0] = 6
    p[1] = 2
  } else if (d < 120) {
    p[0] = 1
    p[1] = 6
    p[2] = 1
  } else if (d < 160) {
    p[1] = 1
    p[2] = 6
    p[3] = 1
  } else if (d < 200) {
    p[2] = 1
    p[3] = 6
    p[4] = 1
  } else if (d < 240) {
    p[3] = 1
    p[4] = 6
    p[5] = 1
  } else if (d < 240) {
    p[4] = 1
    p[5] = 6
    p[6] = 1
  } else if (d < 280) {
    p[5] = 1
    p[6] = 6
    p[7] = 1
  } else if (d < 320) {
    p[6] = 1
    p[7] = 6
    p[8] = 1
  } else if (d < 360) {
    p[7] = 1
    p[8] = 6
    p[9] = 1
  } else if (d < 400) {
    p[8] = 1
    p[9] = 6
    p[10] = 1
  } else {
    p[9] = 2
    p[10] = 6
  }
  let idx = randomChoices(p, 1);
  circle.setAttribute("fill", color[idx]);
}

function addCenterCircle(svg) {
  let svgNS = svg.namespaceURI;
  let circle = document.createElementNS(svgNS, "circle");
  circle.setAttributeNS(null, "cx", 500);
  circle.setAttributeNS(null, "cy", 500);
  circle.setAttributeNS(null, "r", 50);
  circle.setAttribute("fill", '#1c1c1c');

  svg.appendChild(circle);
}

function addRedCircles(svg) {
  let svgNS = svg.namespaceURI;
  for (let i = 0; i < 50; i++) {
    let circle = document.createElementNS(svgNS, "circle");
  
    setCirclePosition(circle, 500, 200);
    setCircleSize(circle);
    circle.setAttribute("fill", '#e46161');
  
    svg.appendChild(circle);
  }
}

function addCircles(svg) {
  let svgNS = svg.namespaceURI;
  for (let i = 0; i < 700; i++) {
    let circle = document.createElementNS(svgNS, "circle");
    setCirclePosition(circle, 500, 140);
    setCircleSize(circle);
    setCircleColor(circle);

    svg.appendChild(circle);
  }

  // addRedCircles(svg)
  // addCenterCircle(svg)
}


svg = initSVG();
addCircles(svg);

document.body.appendChild(svg);

console.log(svg.outerHTML);
