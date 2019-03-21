const hitColour = "red",
  counterColour = "turquoise",
doubleHitColour = "orange";

var cubeWidth = 25,
      xBuffer = 32*cubeWidth,
      yBuffer = 40,
       svgDoc = d3.selectAll("svg");

var updateSizing = function(){
  cubeWidth = window.innerWidth / 50;

  if (typeof inputBytes === 'undefined'){
    xBuffer = 32*cubeWidth;
  } else {
    xBuffer = inputBytes.length * cubeWidth;
  }
};


let strToByteArr = function(str){
  var arr = [];
  for (var i = 0; i < str.length; i++) {
      arr.push(str.charCodeAt(i));
  }

  return arr;
}

let bitArray = function(arr) {
    let output = []; // there is no bit array :(
    for(let i=0; i < arr.length; i++){
      for (let j=0; j < 32; j++) {
          let mask = 1 << j;
          if ((arr[i] & mask) == mask) {
              output.push(1);
          } else {
              output.push(0);
          }
      }
    }

    return output;
}

let appendArray = function(title, backingArray, highlight, yIncrement){
  var items = svgDoc.selectAll("g");

  items.data([title])
       .enter()
       .append("text")
       .attr("x", xBuffer - title.length*5.5) // no clue why it's 5.5, #grafix :D
       .attr("y", yBuffer)
       .text(d => d);

  items.data(backingArray)
       .enter()
       .append("rect")
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth) })
       .attr("y", yBuffer+10)
       .attr("width", cubeWidth)
       .attr("height", cubeWidth)
       .style("fill", highlight);

  items.data(backingArray)
       .enter()
       .append("text")
       .text((d) => d.toString(16)) // mostly this will be bits, but if not hex it
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth) + cubeWidth/4 })
       .attr("y", yBuffer + cubeWidth);

  yBuffer+=yIncrement;
}

let appendText = function(titles, numbers){
  var items = svgDoc.selectAll("g");

  items.data(titles)
       .enter()
       .append("text")
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 })
       .attr("y", yBuffer)
       .text((d) => { return d });

  items.data(numbers)
       .enter()
       .append("rect")
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 })
       .attr("y", yBuffer+10)
       .attr("width", cubeWidth * 4)
       .attr("height", cubeWidth);

  items.data(numbers)
       .enter()
       .append("text")
       .text((d) => {
         var result = d;
         if (typeof(d) === "number") {
          result = d.toString(16); // mostly this will be bits, but if not hex it
         }
         return result;
       })
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 + cubeWidth/4 })
       .attr("y", yBuffer + 1.10*cubeWidth);

  yBuffer+=70;
}

let appendLegend = function(titles, colours){
  var items = svgDoc.selectAll("g");

  items.data(colours)
       .enter()
       .append("rect")
       .style("fill", (d) => { return d })
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth*5 - cubeWidth*3 ) })
       .attr("y", yBuffer)
       .attr("width", 4*cubeWidth)
       .attr("height", cubeWidth);

  items.data(titles)
       .enter()
       .append("text")
       .text((d) => { return d })
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth*7) })
       .attr("y", yBuffer + 0.60*cubeWidth);

  yBuffer+=70;
}

let noop = function(d, i) { return null };

let input = function(hits, doubleHits, pos) {
    return function(d, i) {
        if (doubleHits.includes(i)) {
            return doubleHitColour;
        }
        if (hits.includes(i)) {
            return hitColour;
        }

        if (i == pos) {
            return counterColour;
        }
    };
};

let newHash = function(done){
  inputText = document.getElementById("input").value;
  inputBytes = strToByteArr(inputText);

  $.ajax ({
        url: "/NewHash",
        type: "POST",
        data: JSON.stringify({"data_length":inputBytes.length}),
        dataType: "json",
        contentType: "application/json; charset=utf-8"
  }).fail(function(a,b,c){
    console.log(a,b,c);
    console.log("failed");
  }).done(function(data){
    done(data)
  });
}

let render = function(){
  let dBits = bitArray([inputBytes[ctr]]);
  yBuffer=40;

  svgDoc.html(null);

  appendLegend(["Hits", "Double Hits"], [hitColour, doubleHitColour]);
  appendArray("Input Text", inputText, input(hits, doubleHits, ctr),70);
  appendArray("Input Bytes (hex)", inputBytes, input(hits, doubleHits, ctr),70);
  appendArray("Bits of current selection (d)", dBits.slice(0,8),noop, 70);
  appendArray("Window Array (hex)", fh.rolling_hash.window, noop, 70);

  appendText(["Z Value", "Y Value", "X Value"],
    [fh.rolling_hash.z.toString(10), fh.rolling_hash.y.toString(10), fh.rolling_hash.x.toString(10)]);

  let sig = fh.block_size + ":" + fh.sig1 + ":" + fh.sig2;
  document.getElementById("sig").value = sig;
};

let stepHash = function(){
  ctr++;

  $.ajax ({
        url: "/StepHash",
        type: "POST",
        data: JSON.stringify({"byte":inputBytes[ctr]}),
        dataType: "json",
        contentType: "application/json; charset=utf-8"
  }).fail(function(a,b,c){
    console.log(a,b,c);
    console.log("failed");
  }).done(function(data){
    if (null != data) {
      fh = data // GLOBAL
      if (fh.is_trigger1) {
        hits.push(ctr);
      }
      if (fh.is_trigger2) {
        doubleHits.push(ctr);
      }
      render();
    }
  });
}

let init = function() {

  // GLOBALS
  newHash((data) => {
    fh = data;
    ctr = fh.index;
    hits = [];
    doubleHits = [];

    updateSizing();
    render();
  });
}

init();
