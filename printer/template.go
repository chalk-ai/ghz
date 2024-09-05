package printer

var (
	defaultTmpl = `
Summary:
{{ if .Name }}  Name:		{{ .Name }}
{{ end }}  Count:	{{ .Count }}
  Total:	{{ formatNanoUnit .Total }}
  Slowest:	{{ formatNanoUnit .Slowest }}
  Fastest:	{{ formatNanoUnit .Fastest }}
  Average:	{{ formatNanoUnit .Average }}
  Requests/sec:	{{ formatSeconds .Rps }}

Response time histogram:
{{ histogram .Histogram }}
Latency distribution:{{ range .LatencyDistribution }}
  {{ .Percentage }} % in {{ formatNanoUnit .Latency }} {{ end }}

{{ if gt (len .StatusCodeDist) 0 }}Status code distribution:
{{ formatStatusCode .StatusCodeDist }}{{ end }}
{{ if gt (len .ErrorDist) 0 }}Error distribution:
{{ formatErrorDist .ErrorDist }}{{ end }}
`

	csvTmpl = `
duration (ms),status,error{{ range $i, $v := .Details }}
{{ formatMilli .Latency.Seconds }},{{ .Status }},{{ .Error }}{{ end }}
`

	htmlTmpl = `
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>ghz{{ if .Name }} - {{ .Name }}{{end}}</title>
    <script src="https://d3js.org/d3.v5.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/papaparse@4.5.0/papaparse.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/britecharts@3/dist/bundled/britecharts.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>


    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/britecharts@3/dist/css/britecharts.min.css" type="text/css" /></head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.7.1/css/bulma.min.css" />

  </head>

	<body>

		<section class="section">

    <div class="container" style="display: flex; justify-content: space-between;">
			{{ if .Name }}
			<h1 class="title">{{ .Name }}</h1>
			{{ end }}
			{{ if .Date }}
        <h2 class="subtitle" style="font-size: 2rem; margin-bottom: 0;">{{ formatDate .Date }}</h2>
        <a href="https://chalk.ai" target="_blank">
          <svg version="1.1" id="Layer_1" x="0px" y="0px" viewBox="0 0 577 194.68" xmlns="http://www.w3.org/2000/svg" style="width: 9rem" class="fill-black dark:fill-white h-8"><g><g><g><path d="M81.4,154.59c-24.04,0-39.42-19.96-39.42-42.53c0-22.08,14.88-42.53,38.77-42.53c19.14,0,33.53,10.47,36.64,29.77 H99.39c-1.96-10.47-8.67-16.52-18.32-16.52c-13.25,0-20.28,12.27-20.28,29.28c0,17.18,7.52,29.28,21.1,29.28 c9.49,0,16.36-6.38,18.16-16.85h18.16C115.1,143.8,100.54,154.59,81.4,154.59z"></path><path d="M134.24,152.96V40.09h17.34V87.2h0.33c3.43-11.45,11.94-17.67,23.88-17.67c17.99,0,26.83,12.27,26.83,30.26v53.16 h-17.83v-47.93c0-15.21-5.4-22.08-16.36-22.08c-10.96,0-16.36,7.52-16.36,20.94v49.07H134.24z"></path><path d="M247.6,154.59c-17.01,0-28.95-10.63-28.95-25.52c0-16.85,14.72-23.06,32.88-25.03l20.78-2.13v-2.45 c0-10.96-6.22-16.69-16.52-16.69c-9.98,0-15.86,5.07-16.36,14.23h-17.83c0.65-15.87,13.41-27.48,34.51-27.48 c20.78,0,34.02,11.12,34.02,33.21v50.22h-17.18v-16.03h-0.49C269.03,148.21,259.87,154.59,247.6,154.59z M252.18,141.34 c12.76,0,20.12-8.67,20.12-24.21v-2.78l-17.83,1.96c-10.8,1.14-17.01,5.07-17.01,12.92 C237.46,136.93,244.33,141.34,252.18,141.34z"></path><path d="M311.24,40.09h17.83v112.86h-17.83V40.09z"></path><path d="M368.81,110.26v42.69h-17.83V40.09h17.83v64.78l33.69-33.7h22.08l-35.82,35.99l37.46,45.8h-22.41L368.81,110.26z"></path></g></g><rect x="453.49" y="71.17" width="81.79" height="81.79"></rect></g></svg>
        </a>
			{{ end }}
		</div>

		</div>
		<br />

		<div class="container">
      <nav class="breadcrumb has-bullet-separator" aria-label="breadcrumbs">
        <ul>
          <li>
            <a href="#summary">
              <span class="icon is-small">
                <i class="fas fa-clipboard-list" aria-hidden="true"></i>
              </span>
              <span>Summary</span>
            </a>
          </li>
          <li>
            <a href="#histogram">
              <span class="icon is-small">
                <i class="fas fa-chart-bar" aria-hidden="true"></i>
              </span>
              <span>Histogram</span>
            </a>
          </li>
      	  <li>
            <a href="#rps">
              <span class="icon is-small">
                <i class="fas fa-chart-bar" aria-hidden="true"></i>
              </span>
              <span> RPS </span>
            </a>
          </li>
          <li>
            <a href="#latency">
              <span class="icon is-small">
                <i class="far fa-clock" aria-hidden="true"></i>
              </span>
              <span>Latency Distribution</span>
            </a>
          </li>
          <li>
            <a href="#status">
              <span class="icon is-small">
                <i class="far fa-check-square" aria-hidden="true"></i>
              </span>
              <span>Status Distribution</span>
            </a>
					</li>
					{{ if gt (len .ErrorDist) 0 }}
          <li>
            <a href="#errors">
              <span class="icon is-small">
                <i class="fas fa-exclamation-circle" aria-hidden="true"></i>
              </span>
              <span>Errors</span>
            </a>
					</li>
					{{ end }}
          <li>
            <a href="#data">
              <span class="icon is-small">
                <i class="far fa-file-alt" aria-hidden="true"></i>
              </span>
              <span>Data</span>
            </a>
		  </li>
		  <li>
            <a href="#options">
              <span class="icon is-small">
                <i class="fas fa-cog" aria-hidden="true"></i>
              </span>
              <span>Options</span>
            </a>
          </li>
        </ul>
      </nav>
      <hr />
		</div>

		{{ if gt (len .Tags) 0 }}

			<div class="container">
				<div class="field is-grouped">

				{{ range $tag, $val := .Tags }}

					<div class="control">
						<div class="tags has-addons">
							<span class="tag is-dark">{{ $tag }}</span>
							<span class="tag is-primary">{{ $val }}</span>
						</div>
					</div>

				{{ end }}

				</div>
			</div>
			<br />
		{{ end }}

	  <div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="summary">
							<h3>Summary</h3>
						</a>
						<table class="table">
							<tbody>
								<tr>
									<th>Count</th>
									<td>{{ .Count }}</td>
								</tr>
								<tr>
									<th>Total</th>
									<td>{{ formatNanoUnit .Total }}</td>
								</tr>
								<tr>
									<th>Slowest</th>
								<td>{{ formatNanoUnit .Slowest }}</td>
								</tr>
								<tr>
									<th>Fastest</th>
									<td>{{ formatNanoUnit .Fastest }}</td>
								</tr>
								<tr>
									<th>Average</th>
									<td>{{ formatNanoUnit .Average }}</td>
								</tr>
								<tr>
									<th>Requests / sec</th>
									<td>{{ formatSeconds .Rps }}</td>
								</tr>
							</tbody>
						</table>
					</div>
				</div>
				<div class="column">
					<div class="content">
					</div>
				</div>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="histogram">
					<h3>Histogram</h3>
				</a>
				<p>
					<div class="js-bar-container"></div>
				</p>
			</div>
	  </div>
		<div class="container">
			<div class="content">
				<a name="rps">
					<h3> RPS </h3>
				</a>
				<p>
					<canvas id="js-rps-container"></canvas>
				</p>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="latency">
					<h3>Latency distribution</h3>
				</a>
				<table class="table is-fullwidth">
					<thead>
						<tr>
							{{ range .LatencyDistribution }}
								<th>{{ .Percentage }} %</th>
							{{ end }}
						</tr>
					</thead>
					<tbody>
						<tr>
							{{ range .LatencyDistribution }}
								<td>{{ formatNanoUnit .Latency }}</td>
							{{ end }}
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<br />
		<div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="status">
							<h3>Status distribution</h3>
						</a>
						<table class="table is-hoverable">
							<thead>
								<tr>
									<th>Status</th>
									<th>Count</th>
									<th>% of Total</th>
								</tr>
							</thead>
							<tbody>
							  {{ range $code, $num := .StatusCodeDist }}
									<tr>
									  <td>{{ $code }}</td>
										<td>{{ $num }}</td>
										<td>{{ formatPercent $num $.Count }} %</td>
									</tr>
									{{ end }}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>

			{{ if gt (len .ErrorDist) 0 }}

				<br />
				<div class="container">
					<div class="columns">
						<div class="column is-narrow">
							<div class="content">
								<a name="errors">
									<h3>Errors</h3>
								</a>
								<table class="table is-hoverable">
									<thead>
										<tr>
											<th>Error</th>
											<th>Count</th>
											<th>% of Total</th>
										</tr>
									</thead>
									<tbody>
										{{ range $err, $num := .ErrorDist }}
											<tr>
												<td>{{ $err }}</td>
												<td>{{ $num }}</td>
												<td>{{ formatPercent $num $.Count }} %</td>
											</tr>
											{{ end }}
										</tbody>
									</table>
								</div>
							</div>
						</div>
					</div>

			{{ end }}

			<br />
      <div class="container">
        <div class="columns">
          <div class="column is-narrow">
            <div class="content">
              <a name="data">
                <h3>Data</h3>
              </a>

              <a class="button" id="dlJSON">JSON</a>
              <a class="button" id="dlCSV">CSV</a>
            </div>
          </div>
        </div>
			</div>

			<br />

			<div class="container">
				<div class="content">
					<a name="options">
						<h3>Options</h3>
					</a>
					<article class="message">
						<div class="message-body">
							<pre style="background-color: transparent;">{{ jsonify .Options true }}</pre>
						</div>
					</article>
				</div>
			</div>

			<div class="container">
        <hr />
        <div class="content has-text-centered">
          <p>
            Generated by <strong>ghz</strong>
          </p>
          <a href="https://github.com/bojand/ghz"><i class="icon is-medium fab fa-github"></i></a>
        </div>
      </div>

		</section>

  </body>

  <script>

	const count = {{ .Count }};

	const rawData = {{ jsonify .Details false }};

	const data = [
		{{ range .Histogram }}
			{ name: "{{ .AlternativeMark }}", value: {{ .Count }} },
		{{ end }}
	];

	function createHorizontalBarChart() {
		let barChart = britecharts.bar(),
			tooltip = britecharts.miniTooltip(),
			barContainer = d3.select('.js-bar-container'),
			containerWidth = barContainer.node() ? barContainer.node().getBoundingClientRect().width : false,
			tooltipContainer,
			dataset;

		if (containerWidth) {
			dataset = data;
			barChart
				.isHorizontal(true)
				.isAnimated(true)
				.margin({
					left: 100,
					right: 20,
					top: 20,
					bottom: 20
				})
				.colorSchema(["#B8CEC6","#7BA392","#3F7067","#32645B","#29524A","#264A43","#1C3B34","#0C3129","#06261F","#041D17"])
				.width(containerWidth)
				.yAxisPaddingBetweenChart(20)
				.height(400)
				// .hasPercentage(true)
				.enableLabels(true)
				.labelsNumberFormat('')
				.percentageAxisToMaxRatio(1.3)
				.on('customMouseOver', tooltip.show)
				.on('customMouseMove', tooltip.update)
				.on('customMouseOut', tooltip.hide);

			barChart.orderingFunction(function(a, b) {
				if (a.name.startsWith(">=P99")) {
					return -1;
				}
				if (b.name.startsWith(">=P99")) {
					return 1;
				}

				var nA = a.name.replace(/ms/gi, '');
				var nB = b.name.replace(/ms/gi, '');

				var vA = Number.parseFloat(nA);
				var vB = Number.parseFloat(nB);

				return vB - vA;
			})

			barContainer.datum(dataset).call(barChart);

			tooltipContainer = d3.select('.js-bar-container .bar-chart .metadata-group');
			tooltipContainer.datum([]).call(tooltip);
		}
	}
	function createRPSChart() {
	    let timestamps = rawData.map((item) => {
	        return Date.parse(item["timestamp"])
	    }).sort();

	    let rps = [];
	    let target = timestamps[0] + 1000;
	    let count = 0

	    for (let x of timestamps) {
	        if (x > target + 1000) {
	            rps.push({
	                name: x,
	                value: count
	            })
	            target += 1000
	            count = 0
	        }
	        count += 1
	        if (x == timestamps[-1]) {
	            rps.push({
	                name: x,
	                value: count
	            })
	        }
	    }
	    const ctx = document.getElementById('js-rps-container');

	    new Chart(ctx, {
	            type: 'line',
	            data: {
	                labels: rps.map((item) => (item.name - rps[0].name) / 1000),
	                datasets: [{
	                    label: 'Benchmark RPS',
	                    data: rps.map((item) => item.value),
	                    borderWidth: 1,
                      	borderColor: "#264A43",
	                }]
	            },
	            options: {
	                scales: {
	                    y: {
	                        beginAtZero: true
	                    },
	                    x: {
	                        ticks: {
	                            autoSkip: true,
	                            maxTicksLimit: 20,
	                            callback: function(value) {
	                                return value < 60 ? value + 's' : value / 60 + 'm' + value % 60 + 's'
	                            },
	                        }
	                    }
	                }
	            }
			});
	    }


	    function setJSONDownloadLink() {
	        var filename = "data.json";
	        var btn = document.getElementById('dlJSON');
	        var jsonData = JSON.stringify(rawData)
	        var blob = new Blob([jsonData], {
	            type: 'text/json;charset=utf-8;'
	        });
	        var url = URL.createObjectURL(blob);
	        btn.setAttribute("href", url);
	        btn.setAttribute("download", filename);
	    }

	    function setCSVDownloadLink() {
	        var filename = "data.csv";
	        var btn = document.getElementById('dlCSV');
	        var csv = Papa.unparse(rawData)
	        var blob = new Blob([csv], {
	            type: 'text/csv;charset=utf-8;'
	        });
	        var url = URL.createObjectURL(blob);
	        btn.setAttribute("href", url);
	        btn.setAttribute("download", filename);
	    }

	    createHorizontalBarChart();

	    createRPSChart();

	    setJSONDownloadLink();

	    setCSVDownloadLink();

	</script>

	<script defer src="https://use.fontawesome.com/releases/v5.1.0/js/all.js"></script>

</html>
`
)
