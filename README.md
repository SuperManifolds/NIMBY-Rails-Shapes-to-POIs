# NIMBY ShapeToPOI

A Go web application that converts geographic data files (Shapefiles, KML, KMZ) into NIMBY Rails mod files containing Points of Interest (POI). Available as both a web interface and command-line tool.

## Features

- **Web Interface**: Easy-to-use web application for file conversion
- **Multiple Format Support**: Reads Shapefiles (.shp), KML (.kml), and KMZ (.kmz) files
- **Point Interpolation**: Add intermediate points along lines at configurable distances
- **Label Intervals**: Add name labels at regular intervals along lines
- **Map Preview**: Visual preview of converted POIs on OpenStreetMap
- **Color Customization**: Override colors from files or use original colors
- **POI Splitting**: Automatically splits large datasets into multiple mods (10k POI limit per mod)
- **Nested Geometry Support**: Handles complex nested MultiGeometry structures
- **Multiple File Processing**: Combine data from multiple input files
- **Ready-to-Use Output**: Creates zip files ready for NIMBY Rails import

## Usage

### Web Interface

Visit the deployed application at [your-domain.com] to use the web interface for easy file conversion.

### Local Development

```bash
git clone https://github.com/SuperManifolds/NIMBY-Rails-Shapes-to-POIs
cd NIMBY-Rails-Shapes-to-POIs
make build
./bin/nimby_shapetopoi --server
```

Then open http://localhost:8080 in your browser.

## Deployment

### AWS ECS Deployment

The application is automatically deployed to AWS ECS using Terraform and GitHub Actions.

#### Prerequisites

1. AWS Account with appropriate permissions
2. GitHub repository secrets configured:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

#### Infrastructure Setup

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

The Terraform configuration creates:
- VPC with public/private subnets
- Application Load Balancer
- ECS Cluster with Fargate tasks
- ECR repository for container images
- CloudWatch logging
- Security groups and IAM roles

#### Automatic Deployment

- Push to `main` branch triggers automatic deployment
- Infrastructure changes in `terraform/` directory trigger Terraform workflow
- Application is built as Docker container and deployed to ECS

## Command Line Usage

### Basic Usage

```bash
# Convert a single file
./bin/nimby_shapetopoi depot.kmz

# Convert multiple files
./bin/nimby_shapetopoi *.shp *.kml *.kmz
```

### Advanced Options

```bash
# Specify output file
./bin/nimby_shapetopoi -o my_custom_mod.zip file1.kml file2.kmz

# Use custom mod.txt template
./bin/nimby_shapetopoi -m custom_mod.txt --output combined.zip *.shp

# Combine all options
./bin/nimby_shapetopoi --mod templates/railway.txt --output railway_pois.zip stations.shp tracks.kml
```

## Command Line Options

- `-o, --output <path>`: Output mod zip file path (default: auto-generated)
- `-m, --mod <path>`: Custom mod.txt file to use (default: auto-generated)

## Supported File Formats

### Shapefiles (.shp)
- Points and PolyLines
- Reads "Label" attribute field if present
- Processes associated .dbf, .shx, .prj files

### KML/KMZ Files (.kml, .kmz)
- Points, LineStrings, LinearRings, Polygons
- MultiGeometry (including nested structures)
- Folder hierarchies
- ExtendedData with "Label" field support

## Output Format

The tool generates a zip file containing:
- `mod.txt`: NIMBY Rails mod configuration
- `[name].tsv`: Tab-separated values file with POI data

### TSV Format
```
lon    lat    color       text         font_size  max_lod  transparent  demand  population
10.92  59.98  ff0000ff    Station A    12         10       false        0       0
```

### Mod.txt Format
```ini
[ModMeta]
schema=1
name=depot_mod
author=nimby_shapetopoi
desc=Generated POI layer from geographic files
version=1.0.0

[POILayer]
id = depot_mod_pois
name = depot_mod POIs
tsv = depot_mod.tsv
```

## Development

### Prerequisites
- Go 1.21 or higher
- Make (optional, for build convenience)

### Building
```bash
make build          # Build for current platform
make build-all      # Build for multiple platforms
make build-dev      # Development build with race detector
```

### Testing
```bash
make test           # Run tests
make fmt            # Format code
make lint           # Run linter (requires golangci-lint)
```

### Cleaning
```bash
make clean          # Remove build artifacts and temporary files
```

## Dependencies

- [jonas-p/go-shp](https://github.com/jonas-p/go-shp) - Shapefile reading
- Standard library packages for XML, CSV, ZIP handling

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]