package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	matlas "github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasNetworkPeering_basicAWS(t *testing.T) {
	var peer matlas.Peer

	resourceName := "mongodbatlas_network_peering.test"
	projectID := os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	vpcID := os.Getenv("AWS_VPC_ID")
	vpcCIDRBlock := os.Getenv("AWS_VPC_CIDR_BLOCK")
	awsAccountID := os.Getenv("AWS_ACCOUNT_ID")
	awsRegion := os.Getenv("AWS_REGION")
	providerName := "AWS"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); checkPeeringEnvAWS(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasNetworkPeeringDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigAWS(projectID, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					testAccCheckMongoDBAtlasNetworkPeeringAttributes(&peer, providerName),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", vpcID),
					resource.TestCheckResourceAttr(resourceName, "aws_account_id", awsAccountID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accepter_region_name"},
			},
		},
	})

}

func TestAccResourceMongoDBAtlasNetworkPeering_basicAzure(t *testing.T) {
	var peer matlas.Peer

	resourceName := "mongodbatlas_network_peering.test"
	projectID := os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	directoryID := os.Getenv("AZURE_DIRECTORY_ID")
	subcrptionID := os.Getenv("AZURE_SUBCRIPTION_ID")
	resourceGroupName := os.Getenv("AZURE_RESOURSE_GROUP_NAME")
	vNetName := os.Getenv("AZURE_VNET_NAME")
	providerName := "AZURE"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); checkPeeringEnvAzure(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasNetworkPeeringDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigAzure(projectID, providerName, directoryID, subcrptionID, resourceGroupName, vNetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					testAccCheckMongoDBAtlasNetworkPeeringAttributes(&peer, providerName),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "vnet_name", vNetName),
					resource.TestCheckResourceAttr(resourceName, "azure_directory_id", directoryID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"region_name", "atlas_cidr_block"},
			},
		},
	})
}

func TestAccResourceMongoDBAtlasNetworkPeering_basicGCP(t *testing.T) {
	t.Skip()
	var peer matlas.Peer

	resourceName := "mongodbatlas_network_peering.test"
	projectID := os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	providerName := "GCP"
	gcpProjectID := os.Getenv("GCP_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); checkPeeringEnvGCP(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasNetworkPeeringDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigGCP(projectID, providerName, gcpProjectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					testAccCheckMongoDBAtlasNetworkPeeringAttributes(&peer, providerName),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "azure_directory_id", gcpProjectID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"region_name"},
			},
		},
	})
}

func testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s-%s", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["peer_id"]), nil
	}
}

func testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName string, peer *matlas.Peer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*matlas.Client)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.Attributes["peer_id"] == "" {
			return fmt.Errorf("no ID is set")
		}

		log.Printf("[DEBUG] projectID: %s", rs.Primary.Attributes["project_id"])

		if peerResp, _, err := conn.Peers.Get(context.Background(), rs.Primary.Attributes["project_id"], rs.Primary.Attributes["peer_id"]); err == nil {
			*peer = *peerResp
			peer.ProviderName = rs.Primary.Attributes["provider_name"]
			return nil
		}

		return fmt.Errorf("peer(%s:%s) does not exist", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["peer_id"])
	}
}

func testAccCheckMongoDBAtlasNetworkPeeringAttributes(peer *matlas.Peer, providerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if peer.ProviderName != providerName {
			return fmt.Errorf("bad providerName: %s", peer.RouteTableCIDRBlock)
		}
		return nil
	}
}

func testAccCheckMongoDBAtlasNetworkPeeringDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*matlas.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_network_peering" {
			continue
		}

		// Try to find the peer
		_, _, err := conn.Peers.Get(context.Background(), rs.Primary.Attributes["project_id"], rs.Primary.Attributes["peer_id"])

		if err == nil {
			return fmt.Errorf("peer (%s) still exists", rs.Primary.Attributes["peer_id"])
		}
	}
	return nil
}

func testAccMongoDBAtlasNetworkPeeringConfigAWS(projectID, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = "%[1]s"
			atlas_cidr_block  = "192.168.208.0/21"
			provider_name		  = "%[2]s"
			region_name			  = "%[6]s"
		}

		resource "mongodbatlas_network_peering" "test" {
			accepter_region_name	  = "us-east-1"	
			project_id    			    = "%[1]s"
			container_id            = mongodbatlas_network_container.test.container_id
			provider_name           = "%[2]s"
			route_table_cidr_block  = "%[5]s"
			vpc_id					        = "%[3]s"
			aws_account_id	        = "%[4]s"
		}
	`, projectID, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion)
}

func testAccMongoDBAtlasNetworkPeeringConfigAzure(projectID, providerName, directoryID, subcrptionID, resourceGroupName, vNetName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = "%[1]s"
			atlas_cidr_block  = "192.168.208.0/21"
			provider_name		  = "%[2]s"
			region    			  = "US_EAST_2"
		}

		resource "mongodbatlas_network_peering" "test" {
			project_id   		      = "%[1]s"
			atlas_cidr_block      = "192.168.0.0/21"
			container_id          = mongodbatlas_network_container.test.container_id
			provider_name         = "%[2]s"
			azure_directory_id    = "%[3]s"
			azure_subscription_id = "%[4]s"
			resource_group_name   = "%[5]s"
			vnet_name	            = "%[6]s"
		}
	`, projectID, providerName, directoryID, subcrptionID, resourceGroupName, vNetName)
}

func testAccMongoDBAtlasNetworkPeeringConfigGCP(projectID, providerName, gcpProjectID string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = "%[1]s"
			atlas_cidr_block  = "192.168.208.0/21"
			provider_name		  = "%[2]s"
		}

		resource "mongodbatlas_network_peering" "test" {	
			project_id    	= "%[1]s"
			container_id    = mongodbatlas_network_container.test.container_id
			provider_name   = "%[2]s"
			gcp_project_id  = "%[3]s"
			network_name    = "mongodbatlas_network_container.test.network_name"
		}
	`, projectID, providerName, gcpProjectID)
}
