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
    <title>Chalk Benchmark {{ if .Name }} - {{ .Name }}{{end}}</title>
  	<script src="https://cdn.jsdelivr.net/npm/papaparse@4.5.0/papaparse.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.7.1/css/bulma.min.css" />
    <style>
      .json-key { color: #881391; }
      .json-string { color: #1a1aa6; }
      .json-number { color: #1c00cf; }
      .json-boolean { color: #0d22ff; }
      .json-null { color: #808080; }
    </style>

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
                <i class="fas fa-chart-line" aria-hidden="true"></i>
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
					<canvas id="js-bar-container"></>
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
        <div class="content">
          <a name="data">
            <h3>Data</h3>
          </a>

          <article class="message is-info">
            <div class="message-body">
              <p><strong>Sample:</strong> Showing {{ len .SampleDetails }} of {{ .Count }} requests (first 10K)</p>
              {{ if .Options.CapturePayloads }}
              <p><strong>Note:</strong> Request/response payloads included for sample</p>
              {{ end }}
            </div>
          </article>

          <div class="field is-grouped">
            <p class="control">
              <button class="button is-primary" id="showTableBtn" onclick="loadAndShowTable()">
                <span class="icon">
                  <i class="fas fa-table"></i>
                </span>
                <span>Show Sample Data</span>
              </button>
            </p>
            <p class="control">
              <a class="button" id="dlJSON">
                <span class="icon">
                  <i class="fas fa-download"></i>
                </span>
                <span>JSON</span>
              </a>
            </p>
            <p class="control">
              <a class="button" id="dlCSV">
                <span class="icon">
                  <i class="fas fa-download"></i>
                </span>
                <span>CSV</span>
              </a>
            </p>
          </div>
        </div>

        <div id="dataTableContainer" style="display: none; margin-top: 20px;">
          <div class="field is-grouped">
            <p class="control">
              <input class="input" type="text" id="filterInput" placeholder="Filter by status or error...">
            </p>
            <p class="control">
              <span class="select">
                <select id="pageSizeSelect" onchange="changePageSize()">
                  <option value="10">10 per page</option>
                  <option value="25" selected>25 per page</option>
                  <option value="50">50 per page</option>
                  <option value="100">100 per page</option>
                </select>
              </span>
            </p>
          </div>

          <div style="overflow-x: auto;">
            <table class="table is-striped is-fullwidth is-hoverable">
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Latency</th>
                  <th>Status</th>
                  <th>Error</th>
                  {{ if .Options.CapturePayloads }}
                  <th>Request</th>
                  <th>Response</th>
                  {{ end }}
                </tr>
              </thead>
              <tbody id="dataTableBody">
                <tr><td colspan="6">Loading...</td></tr>
              </tbody>
            </table>
          </div>

          <nav class="pagination is-centered" role="navigation" aria-label="pagination">
            <a class="pagination-previous" id="prevPage" onclick="prevPage()">Previous</a>
            <a class="pagination-next" id="nextPage" onclick="nextPage()">Next</a>
            <ul class="pagination-list" id="paginationList"></ul>
          </nav>
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

	const rawData = {{ jsonify .SampleDetails false }};
	const sampleCount = {{ len .SampleDetails }};
	const totalCount = {{ .Count }};

	const data = [
		{{ range .Histogram }}
			{ name: "{{ .AlternativeMark }}", value: {{ .Count }} },
		{{ end }}
	];

	const rps = {{ jsonify .RPS false }}

	const aggData = {{ jsonify .Aggs false }}

	const createBarChart = () => {
	  const ctx = document.getElementById('js-bar-container');

	  new Chart(ctx, {
		type: 'bar',
		data: {
		  labels: data.map((item) => item.name),
		  datasets: [{
			label: 'Latency Bucket',
			data: data.map((item) => item.value),
			borderWidth: 1,
			barThickness: 20,
			backgroundColor: ["#B8CEC6", "#7BA392", "#3F7067", "#32645B", "#29524A", "#264A43", "#1C3B34", "#0C3129", "#06261F", "#041D17"]
		  }]
		},
		options: {
		  aspectRatio: 3,
		  plugins: {
			legend: {
			  display: false
			},
		  },
		  tooltips: {
			callbacks: {
			  label: function(tooltipItem) {
				return tooltipItem.yLabel;
			  }
			}
		  },
		  indexAxis: 'y',
		  scales: {
			y: {
			  ticks: {
				callback: function(value) {
				  if (value == data.length - 1) {
					return data[value].name.split(' ')[0]
				  } else {
					return parseFloat(data[value].name).toFixed(1) + 'ms'
				  }
				}
			  }
			}
		  }
		},
	  });
	}

	const createRPSChart = () => {
	  const ctx = document.getElementById('js-rps-container');

	  new Chart(ctx, {
		type: 'line',
		data: {
		  labels: rps.map((item) => (item.x / 1000)),
		  datasets: [
			{
			  label: 'RPS',
			  data: rps,
			  borderWidth: 2,
			  borderColor: "rgba(38, 74, 67, 1)",
			  backgroundColor: "rgba(38, 74, 67, 1)",
			  pointRadius: 0,
			  tension: 0.3,
			},
			{{ if .P50}}
			{
			  label: 'P50',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p50 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(112, 206, 151)",
			  backgroundColor: "rgb(112, 206, 151, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  lineWidth: 2,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P95}}
			{
			  label: 'P95',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p95 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(129, 206, 234)",
			  backgroundColor: "rgb(129, 206, 234, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  lineWidth: 2,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P99 }}
			{
			  label: 'P99',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p99 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(249, 149, 127)",
			  backgroundColor: "rgb(249, 149, 127, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  yAxisID: "y2"
			},
			{{ end }}
			{{ if .P99_9 }}
			{
			  label: 'P99.9',
			  data: aggData.map((item) => ({ x: item.x / 1000, y: item.y.p99_9 / 1e6 })),
			  borderWidth: 1.5,
			  borderColor: "rgb(220, 100, 100)",
			  backgroundColor: "rgb(220, 100, 100, .8)",
			  tension: 0.3,
			  pointRadius: 0,
			  yAxisID: "y2"
			}
			{{ end }}
		  ]
		},
		options: {
		  plugins: {
			legend: {
				align: "end",
				labels: {
				  boxWidth: 20
				}
			  }
		  },
		  aspectRatio: 2.5,
		  layout: {
			padding: {
			  left: 30
			}
		  },
		  scales: {
			y:
			{
			  beginAtZero: true,
			  type: 'linear',
			  position: 'left',
			  stack: "demo",
			  stackWeight: 2,
			},
			y2: {
			  type: 'linear',
			  position: 'left',
			  offset: true,
			  stack: "demo",
			  stackWeight: 3,
			  ticks: {
          callback: function(value) {
            return value + 'ms'
          },
			  },
			},
			x: {
          ticks: {
            callback: function(value) {
              return value < 60 ? value + 's' : Math.floor(value / 60) + 'm' + value % 60 + 's'
            },
            autoSkip: true,
            maxTicksLimit: 20,
          }
        }
		  }
		}
	  });
	}

	function downloadJSON() {
	  var btn = document.getElementById('dlJSON');
	  btn.classList.add('is-loading');

	  try {
		var jsonData = JSON.stringify(rawData, null, 2);
		var blob = new Blob([jsonData], {
		  type: 'text/json;charset=utf-8;'
		});
		var url = URL.createObjectURL(blob);
		var a = document.createElement('a');
		a.href = url;
		a.download = 'data.json';
		a.click();
		URL.revokeObjectURL(url);
	  } catch (error) {
		alert('Failed to download JSON: ' + error.message);
	  } finally {
		btn.classList.remove('is-loading');
	  }
	}

	function downloadCSV() {
	  let btn = document.getElementById('dlCSV');
	  btn.classList.add('is-loading');

	  try {
		let csv = Papa.unparse(rawData);
		let blob = new Blob([csv], {
		  type: 'text/csv;charset=utf-8;'
		});
		let url = URL.createObjectURL(blob);
		let a = document.createElement('a');
		a.href = url;
		a.download = 'data.csv';
		a.click();
		URL.revokeObjectURL(url);
	  } catch (error) {
		alert('Failed to download CSV: ' + error.message);
	  } finally {
		btn.classList.remove('is-loading');
	  }
	}

	document.getElementById('dlJSON').addEventListener('click', downloadJSON);
	document.getElementById('dlCSV').addEventListener('click', downloadCSV);

	// Table display state
	let tableData = [];
	let filteredData = [];
	let currentPage = 1;
	let pageSize = 25;

	function loadAndShowTable() {
	  const btn = document.getElementById('showTableBtn');
	  btn.classList.add('is-loading');
	  btn.disabled = true;

	  try {
		console.log('Loading data...');
		tableData = rawData;
		filteredData = tableData;

		console.log('Sample data row:', tableData.length > 0 ? tableData[0] : 'No data');
		console.log('Capture payloads:', {{ .Options.CapturePayloads }});

		if (tableData.length === 0) {
		  alert('No data found in sample. This might indicate an issue with data collection.');
		  btn.classList.remove('is-loading');
		  btn.disabled = false;
		  return;
		}

		renderTable();
		document.getElementById('dataTableContainer').style.display = 'block';
		btn.style.display = 'none';

		// Add filter listener with debounce
		let filterTimeout;
		document.getElementById('filterInput').addEventListener('input', (e) => {
		  clearTimeout(filterTimeout);
		  filterTimeout = setTimeout(() => {
			filterTable(e.target.value);
		  }, 300);
		});
	  } catch (error) {
		console.error('Error loading table:', error);
		btn.classList.remove('is-loading');
		btn.disabled = false;
		alert('Failed to load table data: ' + error.message);
	  }
	}

	function filterTable(searchTerm) {
	  if (!searchTerm) {
		filteredData = tableData;
	  } else {
		const term = searchTerm.toLowerCase();
		filteredData = tableData.filter(row =>
		  (row.status && row.status.toLowerCase().includes(term)) ||
		  (row.error && row.error.toLowerCase().includes(term))
		);
	  }
	  currentPage = 1;
	  renderTable();
	}

	function changePageSize() {
	  pageSize = parseInt(document.getElementById('pageSizeSelect').value);
	  currentPage = 1;
	  renderTable();
	}

	function prevPage() {
	  if (currentPage > 1) {
		currentPage--;
		renderTable();
	  }
	}

	function nextPage() {
	  const totalPages = Math.ceil(filteredData.length / pageSize);
	  if (currentPage < totalPages) {
		currentPage++;
		renderTable();
	  }
	}

	function goToPage(page) {
	  currentPage = page;
	  renderTable();
	}

	function formatLatency(latency) {
	  if (latency !== undefined && latency !== null) {
		// latency is in nanoseconds
		const ms = latency / 1000000;
		return ms.toFixed(2) + ' ms';
	  }
	  return '-';
	}

	function formatTimestamp(timestamp) {
	  if (timestamp) {
		const d = new Date(timestamp);
		return d.toLocaleString();
	  }
	  return '-';
	}

	function syntaxHighlight(json) {
	  json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
		var cls = 'json-number';
		if (/^"/.test(match)) {
		  if (/:$/.test(match)) {
			cls = 'json-key';
		  } else {
			cls = 'json-string';
		  }
		} else if (/true|false/.test(match)) {
		  cls = 'json-boolean';
		} else if (/null/.test(match)) {
		  cls = 'json-null';
		}
		return '<span class="' + cls + '">' + match + '</span>';
	  });
	}

	function formatPayload(payload) {
	  if (!payload) return '-';

	  try {
		// Try to parse as JSON (protobuf messages are serialized as JSON)
		const parsed = JSON.parse(payload);
		const formatted = JSON.stringify(parsed, null, 2);

		// Create a compact summary for protobuf messages
		const summary = createPayloadSummary(parsed);

		// If it's very short, just show it all with highlighting
		if (formatted.length <= 100) {
		  return '<pre style="margin: 0; max-width: 400px; overflow-x: auto; white-space: pre-wrap; word-wrap: break-word;">' + syntaxHighlight(formatted) + '</pre>';
		}

		// Otherwise show summary with expand button
		const id = 'payload-' + Math.random().toString(36).substring(7);

		return '<div>' +
		  '<div id="' + id + '-summary" style="margin: 0; max-width: 400px; display: block;">' + summary + '</div>' +
		  '<pre id="' + id + '-full" style="margin: 0; max-width: 400px; overflow-x: auto; white-space: pre-wrap; word-wrap: break-word; display: none;">' + syntaxHighlight(formatted) + '</pre>' +
		  '<a href="#" onclick="togglePayload(\'' + id + '\'); return false;" style="font-size: 0.75rem; color: #3273dc;">Expand</a>' +
		  '</div>';
	  } catch (e) {
		// Not valid JSON, show as plain text
		if (payload.length <= 100) return payload;
		return payload.substring(0, 100) + '...';
	  }
	}

	function createPayloadSummary(obj, maxDepth = 2) {
	  if (typeof obj !== 'object' || obj === null) {
		return String(obj);
	  }

	  if (Array.isArray(obj)) {
		return '[' + obj.length + ' items]';
	  }

	  // Create a compact representation of the object
	  const keys = Object.keys(obj);
	  if (keys.length === 0) return '{}';

	  const parts = [];
	  for (let i = 0; i < Math.min(keys.length, 3); i++) {
		const key = keys[i];
		const val = obj[key];

		if (typeof val === 'object' && val !== null) {
		  if (Array.isArray(val)) {
			parts.push(key + ': [' + val.length + ']');
		  } else {
			parts.push(key + ': {...}');
		  }
		} else {
		  const valStr = String(val);
		  parts.push(key + ': ' + (valStr.length > 20 ? valStr.substring(0, 20) + '...' : valStr));
		}
	  }

	  let result = '{ ' + parts.join(', ');
	  if (keys.length > 3) {
		result += ', ... +' + (keys.length - 3) + ' fields';
	  }
	  result += ' }';

	  return result;
	}

	function togglePayload(id) {
	  const summary = document.getElementById(id + '-summary');
	  const full = document.getElementById(id + '-full');
	  const link = event.target;

	  if (summary.style.display === 'none') {
		summary.style.display = 'block';
		full.style.display = 'none';
		link.textContent = 'Expand';
	  } else {
		summary.style.display = 'none';
		full.style.display = 'block';
		link.textContent = 'Collapse';
	  }
	}

	function renderTable() {
	  const tbody = document.getElementById('dataTableBody');
	  const totalPages = Math.ceil(filteredData.length / pageSize);
	  const startIdx = (currentPage - 1) * pageSize;
	  const endIdx = Math.min(startIdx + pageSize, filteredData.length);

	  // Clear tbody
	  tbody.innerHTML = '';

	  // Show message if no data
	  if (filteredData.length === 0) {
		const tr = document.createElement('tr');
		const hasPayloads = {{ .Options.CapturePayloads }};
		const colspan = hasPayloads ? 6 : 4;
		tr.innerHTML = '<td colspan="' + colspan + '" style="text-align: center; padding: 20px;">No data found</td>';
		tbody.appendChild(tr);
		return;
	  }

	  // Render rows for current page
	  for (let i = startIdx; i < endIdx; i++) {
		const row = filteredData[i];
		const tr = document.createElement('tr');

		const hasPayloads = {{ .Options.CapturePayloads }};

		if (i === startIdx && hasPayloads) {
		  console.log('First row payloads:', {
			request: row.requestPayload ? row.requestPayload.substring(0, 100) : 'MISSING',
			response: row.responsePayload ? row.responsePayload.substring(0, 100) : 'MISSING'
		  });
		}

		let html = '<td>' + formatTimestamp(row.timestamp) + '</td>' +
		  '<td>' + formatLatency(row.latency) + '</td>' +
		  '<td><span class="tag ' + (row.status === 'OK' ? 'is-success' : 'is-danger') + '">' + (row.status || '-') + '</span></td>' +
		  '<td>' + (row.error || '-') + '</td>';

		if (hasPayloads) {
		  html += '<td style="font-size: 0.75rem;">' + formatPayload(row.requestPayload) + '</td>' +
			'<td style="font-size: 0.75rem;">' + formatPayload(row.responsePayload) + '</td>';
		}

		tr.innerHTML = html;
		tbody.appendChild(tr);
	  }

	  // Update pagination
	  renderPagination(totalPages);

	  // Update prev/next buttons
	  document.getElementById('prevPage').disabled = currentPage === 1;
	  document.getElementById('nextPage').disabled = currentPage === totalPages;
	}

	function renderPagination(totalPages) {
	  const paginationList = document.getElementById('paginationList');
	  paginationList.innerHTML = '';

	  // Show max 7 page numbers
	  let startPage = Math.max(1, currentPage - 3);
	  let endPage = Math.min(totalPages, startPage + 6);

	  if (endPage - startPage < 6) {
		startPage = Math.max(1, endPage - 6);
	  }

	  for (let i = startPage; i <= endPage; i++) {
		const li = document.createElement('li');
		const a = document.createElement('a');
		a.className = 'pagination-link' + (i === currentPage ? ' is-current' : '');
		a.textContent = i;
		a.onclick = () => goToPage(i);
		li.appendChild(a);
		paginationList.appendChild(li);
	  }
	}

	createBarChart();

	createRPSChart();

	setCSVDownloadLink();
	setJSONDownloadLink();
	</script>
	<script defer src="https://use.fontawesome.com/releases/v5.1.0/js/all.js"></script>
</html>
`
)
