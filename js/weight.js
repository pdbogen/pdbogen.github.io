var rawData = new Array();
var dailyNotes = new Array();
var dailyData = new Array();
var fiveDayData = new Array();
var thirtyDayData = new Array();
var series = new Array();
var xLabel;

function addDailyData( unixtime, value, note ) {
	var d = new Date( unixtime*1000 );
	d.setHours( 12 );
	d.setMinutes( 0 );
	d.setSeconds( 0 );
	d.setMilliseconds( 0 );
	if( dailyData[ d.valueOf() ] ) {
		dailyData[ d.valueOf() ].sum += value;
		dailyData[ d.valueOf() ].count++;
		if( note ) {
			if( dailyData[ d.valueOf() ].note ) {
				dailyData[ d.valueOf() ].note += "\n" + note;
			} else {
				dailyData[ d.valueOf() ].note = note;
			}
		}
	} else {
		if( note ) {
			dailyData[ d.valueOf() ] = { sum: value, count: 1, note: note };
		} else {
			dailyData[ d.valueOf() ] = { sum: value, count: 1 };
		}
	}
}

function finishDailyData() {
	// First pass, calculate daily averages
	for( var d in dailyData ) {
		if( dailyData[d].note ) {
			var day = d / 86400 / 1000;
			if( dailyNotes[ day ] ) {
				dailyNotes[ day ] += "\n" + dailyData[d].note;
			} else {
				dailyNotes[ day ] = dailyData[d].note;
			}
		}
		dailyData[d] = dailyData[d].sum / dailyData[d].count;
	}

	// Second pass, calculate 5- and 30-day moving averages
	for( var d in dailyData ) {
		var count = 0;
		var sum = 0;
		for( var fd = d - 86400*4*1000; fd <= d; fd += 86400*1000 ) {
			if( dailyData[ fd ] ) {
				count++;
				sum += dailyData[fd];
			}
		}
		fiveDayData.push( { x: Number( d ) / 1000, y: sum/count } );

		count = 0; sum = 0;
		for( var fd = d - 86400*29*1000; fd <= d; fd += 86400*1000 ) {
			if( dailyData[fd] ) {
				count++;
				sum += dailyData[fd];
			}
		}
		thirtyDayData.push( { x: Number( d ) / 1000, y: sum/count } );
	}
}

var x_axis;

function loadData( payload ) {
	var s = "";
	var cells = payload.feed.entry;
	var titleRegex = /^([a-z])([0-9]+)$/i;
	var x, y;
	var rows = new Array();
	for( var i in cells ) {
		var cell = cells[i];
		var row = new Number( cell[ "gs$cell" ][ "row" ] ) - 1;
		var col = new Number( cell[ "gs$cell" ][ "col" ] ) - 1;
		var rawValue = new Number( cell[ "gs$cell" ][ "numericValue" ] );
		var renderValue = cell[ "gs$cell" ][ "$t" ];
		if( col == 0 ) { // New row, so start a new row in the data table
			rawData[ row ] = new Array();
		}
		if( row == 0 ) { // first row, list of labels
			rawData[ 0 ][ col ] = renderValue;
		} else {
			// Add this to the list of rows to tabulate
			rows[ row ] = 1;

			if( col == 0 ) { // special processing for the time field: convert to epoch time
				rawData[ row ][ col ] = (rawValue.valueOf() - 25569 + 7/24) * 86400;
			} else if( col == 1 ) {
				rawData[ row ][ col ] = rawValue.valueOf();
			} else if( col == 5 ) {
				rawData[ row ][ col ] = cell[ "gs$cell" ][ "inputValue" ];
			}
		}
	}

	for( var row in rows ) {
		addDailyData( rawData[ row ][ 0 ], rawData[ row ][ 1 ], rawData[ row ][ 5 ] );
	}

	finishDailyData();
	for( var s = 1; s < rawData[0].length; s++ ) {
		series[s] = {
			title: rawData[0][s],
			rawData: new Array(),
		};
		for( var i = 1; i < rawData.length; i++ ) {
			series[s].rawData.push( { x: rawData[i][0], y: rawData[i][s] } );
		}
	}
	xLabel = rawData[0][0];
	var palette = new Rickshaw.Color.Palette( { scheme: 'munin' } );
	var graph = new Rickshaw.Graph( {
		element: document.querySelector( "#chart" ),
		height: window.innerHeight*.9,
		min: 'auto',
		renderer: 'multi',
		interpolation: 'basis',
		series: [ {
			name: series[1].title,
			color: palette.color(),
			renderer: 'scatterplot',
			data: series[1].rawData,
		}, {
			name: "5-Day Average",
			color: palette.color(),
			data: fiveDayData,
			renderer: 'line'
		}, {
			name: "30-Day Average",
			color: palette.color(),
			data: thirtyDayData,
			renderer: 'line',
		} ],
	});
	x_axis = new Rickshaw.Graph.Axis.Time( { graph: graph } );
	var y_axis = new Rickshaw.Graph.Axis.Y( {
		graph: graph,
		orientation: 'left',
		element: document.getElementById( "y_axis" ),
	} );
	var legend = new Rickshaw.Graph.Legend( {
		graph: graph,
		element: document.getElementById( "legend" ),
	} );
	var highlight = new Rickshaw.Graph.Behavior.Series.Highlight( {
		graph: graph,
		legend: legend
	} );
	var hoverDetail = new Rickshaw.Graph.HoverDetail( {
		graph: graph,
		xFormatter: function(x) {
			var d = new Date(x*1000);
			return d.getFullYear() + "-" +
				(d.getMonth()<9?"0":"") +
				(d.getMonth()+1) + "-" +
				(d.getDate()<10?"0":"") +
				d.getDate() + " " +
				(d.getHours()<10?"0":"") +
				d.getHours() + ":" +
				(d.getMinutes()<10?"0":"") +
				d.getMinutes();
		}
	} );

	var preview = new Rickshaw.Graph.RangeSlider( {
		graph: graph,
		element: document.getElementById('preview')
	} );


	var annotator = new Rickshaw.Graph.Annotate( {
		graph: graph,
		element: document.getElementById( 'timeline' )
	} );

	for( var day in dailyNotes ) {
		if( dailyNotes[ day ] ) {
			annotator.add( day*86400, dailyNotes[ day ] );
		}
	}

	annotator.update();
	graph.render();
}
