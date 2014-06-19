var data = new Array();
var series = new Array();
var xLabel;
function loadData( payload ) {
	var s = "";
	var cells = payload.feed.entry;
	var titleRegex = /^([a-z])([0-9]+)$/i;
	var x, y;
	for( var i in cells ) {
		var cell = cells[i];
		var row = new Number( cell[ "gs$cell" ][ "row" ] ) - 1;
		var col = new Number( cell[ "gs$cell" ][ "col" ] ) - 1;
		var rawValue = new Number( cell[ "gs$cell" ][ "numericValue" ] );
		var renderValue = cell[ "gs$cell" ][ "$t" ];
		if( col == 0 ) { // New row, so start a new row in the data table
			data[ row ] = new Array();
		}
		if( row == 0 ) { // first row, list of labels
			data[ 0 ][ col ] = renderValue;
		} else {
			if( col == 0 ) { // special processing for the time field: convert to epoch time
				data[ row ][ col ] = (rawValue.valueOf() - 25569 + 7/24) * 86400;
			} else {
				data[ row ][ col ] = rawValue.valueOf();
			}
		}
	}
	for( var s = 1; s < data[0].length; s++ ) {
		series[s] = {
			title: data[0][s],
			data: new Array(),
		};
		for( var i = 1; i < data.length; i++ ) {
			series[s].data.push( { x: data[i][0], y: data[i][s] } );
		}
	}
	xLabel = data[0][0];
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
			data: series[1].data,
		}, {
			name: series[3].title,
			color: palette.color(),
			data: series[3].data,
			renderer: 'line'
		}, {
			name: series[4].title,
			color: palette.color(),
			data: series[4].data,
			renderer: 'line',
		} ],
	});
	var x_axis = new Rickshaw.Graph.Axis.Time( { graph: graph } );
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
		element: document.getElementById('preview'),
	} );

	graph.render();
}
