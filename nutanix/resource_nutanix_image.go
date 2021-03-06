package nutanix

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-nutanix/client/v3"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	// ImageKind Represents kind of resource
	ImageKind = "image"
	// DELETED ...
	DELETED = "DELETED"
	// ERROR ..
	ERROR = "ERROR"
)

func resourceNutanixImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceNutanixImageCreate,
		Read:   resourceNutanixImageRead,
		Update: resourceNutanixImageUpdate,
		Delete: resourceNutanixImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"api_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"last_update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"uuid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spec_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spec_hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"categories": categoriesSchema(),
			"owner_reference": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"project_reference": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone_reference": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Required: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"cluster_reference": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Required: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"retrieval_uri_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"image_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"checksum_algorithm": {
							Type:     schema.TypeString,
							Required: true,
						},
						"checksum_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"source_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"source_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"product_version": {
							Type:     schema.TypeString,
							Required: true,
						},
						"product_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceNutanixImageCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating Image: %s", d.Get("name").(string))

	conn := meta.(*Client).API

	request := &v3.ImageIntentInput{}
	spec := &v3.Image{}
	metadata := &v3.Metadata{}
	image := &v3.ImageResources{}

	n, nok := d.GetOk("name")
	desc, descok := d.GetOk("description")

	_, iok := d.GetOk("source_uri")
	_, pok := d.GetOk("source_path")

	// if both path and uri are provided, return an error
	if iok && pok {
		return errors.New("both source_uri and source_path provided")
	}

	// Read Arguments and set request values
	if !nok {
		return fmt.Errorf("please provide the required attribute name")
	}

	if err := getMetadataAttributes(d, metadata, "image"); err != nil {
		return err
	}

	if descok {
		spec.Description = utils.StringPtr(desc.(string))
	}

	if err := getImageResource(d, image); err != nil {
		return err
	}

	spec.Name = utils.StringPtr(n.(string))
	spec.Resources = image

	request.Metadata = metadata
	request.Spec = spec

	imageUUID, err := resourceNutanixImageExists(conn, n.(string))
	if err != nil {
		return fmt.Errorf("failed to read image with name(%s): %+v", n.(string), err)
	}

	if imageUUID != nil {
		return fmt.Errorf("image already exists with name %s  in the given cluster, UUID %s", d.Get("name").(string), *imageUUID)
	}

	// Make request to the API
	resp, err := conn.V3.CreateImage(request)
	if err != nil {
		return err
	}

	UUID := *resp.Metadata.UUID
	// set terraform state
	d.SetId(UUID)

	taskUUID := resp.Status.ExecutionContext.TaskUUID.(string)

	// Wait for the Image to be available
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"QUEUED", "RUNNING"},
		Target:     []string{"SUCCEEDED"},
		Refresh:    taskStateRefreshFunc(conn, taskUUID),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, errw := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for image (%s) to create: %s", d.Id(), errw)
	}

	// if we need to upload an image, we do it now
	if pok {
		path := d.Get("source_path")

		err = conn.V3.UploadImage(UUID, path.(string))
		if err != nil {

			delErr := resourceNutanixImageDelete(d, meta)
			if delErr != nil {
				return delErr
			}

			return fmt.Errorf("failed uploading image: %s", err)
		}
	}

	return resourceNutanixImageRead(d, meta)
}

func resourceNutanixImageRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Image: %s", d.Get("name").(string))

	// Get client connection
	conn := meta.(*Client).API
	uuid := d.Id()

	// Make request to the API
	resp, err := conn.V3.GetImage(uuid)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "ENTITY_NOT_FOUND") {
			d.SetId("")
		}
		return fmt.Errorf("error reading image UUID (%s) with error %s", uuid, err)
	}

	m, c := setRSEntityMetadata(resp.Metadata)

	if err = d.Set("metadata", m); err != nil {
		return fmt.Errorf("error setting metadata for image UUID(%s), %s", d.Id(), err)
	}
	if err = d.Set("categories", c); err != nil {
		return fmt.Errorf("error setting categories for image UUID(%s), %s", d.Id(), err)
	}

	if err = d.Set("owner_reference", flattenReferenceValues(resp.Metadata.OwnerReference)); err != nil {
		return fmt.Errorf("error setting owner_reference for image UUID(%s), %s", d.Id(), err)
	}
	d.Set("api_version", utils.StringValue(resp.APIVersion))
	d.Set("name", utils.StringValue(resp.Status.Name))
	d.Set("description", utils.StringValue(resp.Status.Description))

	if err = d.Set("availability_zone_reference", flattenReferenceValues(resp.Status.AvailabilityZoneReference)); err != nil {
		return fmt.Errorf("error setting owner_reference for image UUID(%s), %s", d.Id(), err)
	}
	if err = d.Set("cluster_reference", getClusterReferenceValues(resp.Status.ClusterReference)); err != nil {
		return fmt.Errorf("error setting cluster_reference for image UUID(%s), %s", d.Id(), err)
	}

	if err = d.Set("state", resp.Status.State); err != nil {
		return fmt.Errorf("error setting state for image UUID(%s), %s", d.Id(), err)
	}

	if err = d.Set("image_type", resp.Status.Resources.ImageType); err != nil {
		return fmt.Errorf("error setting image_type for image UUID(%s), %s", d.Id(), err)
	}

	if err = d.Set("source_uri", resp.Status.Resources.SourceURI); err != nil {
		return fmt.Errorf("error setting source_uri for image UUID(%s), %s", d.Id(), err)
	}

	if err = d.Set("size_bytes", resp.Status.Resources.SizeBytes); err != nil {
		return fmt.Errorf("error setting size_bytes for image UUID(%s), %s", d.Id(), err)
	}

	checksum := make(map[string]string)
	if resp.Status.Resources.Checksum != nil {
		checksum["checksum_algorithm"] = utils.StringValue(resp.Status.Resources.Checksum.ChecksumAlgorithm)
		checksum["checksum_value"] = utils.StringValue(resp.Status.Resources.Checksum.ChecksumValue)
	}

	if err = d.Set("checksum", checksum); err != nil {
		return fmt.Errorf("error setting checksum for image UUID(%s), %s", d.Id(), err)
	}

	version := make(map[string]string)
	if resp.Status.Resources.Version != nil {
		version["product_version"] = utils.StringValue(resp.Status.Resources.Version.ProductVersion)
		version["product_name"] = utils.StringValue(resp.Status.Resources.Version.ProductName)
	}

	if err = d.Set("version", version); err != nil {
		return fmt.Errorf("error setting version for image UUID(%s), %s", d.Id(), err)
	}

	uriList := make([]string, 0, len(resp.Status.Resources.RetrievalURIList))
	for _, uri := range resp.Status.Resources.RetrievalURIList {
		uriList = append(uriList, utils.StringValue(uri))
	}

	if err = d.Set("retrieval_uri_list", uriList); err != nil {
		return fmt.Errorf("error setting retrieval_uri_list for image UUID(%s), %s", d.Id(), err)
	}

	return nil
}

func resourceNutanixImageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Client).API

	// get state
	request := &v3.ImageIntentInput{}
	metadata := &v3.Metadata{}
	spec := &v3.Image{}
	res := &v3.ImageResources{}

	response, err := conn.V3.GetImage(d.Id())

	if err != nil {
		if strings.Contains(fmt.Sprint(err), "ENTITY_NOT_FOUND") {
			d.SetId("")
		}
		return err
	}

	if response.Metadata != nil {
		metadata = response.Metadata
	}

	if response.Spec != nil {
		spec = response.Spec

		if response.Spec.Resources != nil {
			res = response.Spec.Resources
		}
	}

	if d.HasChange("categories") {
		catl := d.Get("categories").(map[string]interface{})
		metadata.Categories = expandCategories(catl)
	}

	if d.HasChange("owner_reference") {
		or := d.Get("owner_reference").(map[string]interface{})
		metadata.OwnerReference = validateRef(or)
	}

	if d.HasChange("project_reference") {
		pr := d.Get("project_reference").(map[string]interface{})
		metadata.ProjectReference = validateRef(pr)
	}

	if d.HasChange("name") {
		spec.Name = utils.StringPtr(d.Get("name").(string))
	}
	if d.HasChange("description") {
		spec.Description = utils.StringPtr(d.Get("description").(string))
	}

	if d.HasChange("source_uri") || d.HasChange("checksum") {
		if err := getImageResource(d, res); err != nil {
			return err
		}
		spec.Resources = res
	}

	request.Metadata = metadata
	request.Spec = spec

	resp, errUpdate := conn.V3.UpdateImage(d.Id(), request)

	if errUpdate != nil {
		return fmt.Errorf("error updating image(%s) %s", d.Id(), errUpdate)
	}

	taskUUID := resp.Status.ExecutionContext.TaskUUID.(string)

	// Wait for the Image to be available
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"QUEUED", "RUNNING"},
		Target:     []string{"SUCCEEDED"},
		Refresh:    taskStateRefreshFunc(conn, taskUUID),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for image (%s) to update: %s", d.Id(), err)
	}

	return resourceNutanixImageRead(d, meta)
}

func resourceNutanixImageDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting Image: %s", d.Get("name").(string))

	conn := meta.(*Client).API
	UUID := d.Id()

	resp, err := conn.V3.DeleteImage(UUID)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "ENTITY_NOT_FOUND") {
			d.SetId("")
		}
		return err
	}

	taskUUID := resp.Status.ExecutionContext.TaskUUID.(string)

	// Wait for the Image to be available
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"QUEUED", "RUNNING"},
		Target:     []string{"SUCCEEDED"},
		Refresh:    taskStateRefreshFunc(conn, taskUUID),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for image (%s) to delete: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func getImageResource(d *schema.ResourceData, image *v3.ImageResources) error {
	cs, csok := d.GetOk("checksum")
	checks := &v3.Checksum{}

	if su, suok := d.GetOk("source_uri"); suok {
		ext := filepath.Ext(su.(string))
		if ext == ".qcow2" {
			image.ImageType = utils.StringPtr("DISK_IMAGE")
		} else if ext == ".iso" {
			image.ImageType = utils.StringPtr("ISO_IMAGE")
		} else {
			// By default assuming the image to be raw disk image.
			image.ImageType = utils.StringPtr("DISK_IMAGE")
		}
		// set source uri
		image.SourceURI = utils.StringPtr(su.(string))
	}

	if csok {
		checksum := cs.(map[string]interface{})
		ca, caok := checksum["checksum_algorithm"]
		cv, cvok := checksum["checksum_value"]

		if caok {
			if ca.(string) == "" {
				return fmt.Errorf("'checksum_algorithm' is not given")
			}
			checks.ChecksumAlgorithm = utils.StringPtr(ca.(string))
		}
		if cvok {
			if cv.(string) == "" {
				return fmt.Errorf("'checksum_value' is not given")
			}
			checks.ChecksumValue = utils.StringPtr(cv.(string))
		}
		image.Checksum = checks
	}

	return nil
}

func resourceNutanixImageExists(conn *v3.Client, name string) (*string, error) {
	log.Printf("[DEBUG] Get Image Existence : %s", name)

	imageEntities := &v3.DSMetadata{}
	var imageUUID *string

	imageList, err := conn.V3.ListImage(imageEntities)

	if err != nil {
		return nil, err
	}

	for _, image := range imageList.Entities {
		if image.Status.Name == utils.StringPtr(name) {
			imageUUID = image.Metadata.UUID
		}
	}
	return imageUUID, nil
}
