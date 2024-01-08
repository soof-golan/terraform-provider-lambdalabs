// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"terraform-provider-lambdalabs/pgk/lambdalabs"
	"testing"
)

var _ lambdalabs.ServerInterface = &MockSSHKeyServer{}

type MockSSHKeyServer struct {
}

func (m MockSSHKeyServer) ListFileSystems(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) LaunchInstance(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) RestartInstance(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) TerminateInstance(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) InstanceTypes(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) ListInstances(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) GetInstance(c *gin.Context, id lambdalabs.InstanceId) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) ListSSHKeys(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockSSHKeyServer) AddSSHKey(c *gin.Context) {

}

func (m MockSSHKeyServer) DeleteSSHKey(c *gin.Context, id lambdalabs.SshKeyId) {
	c.JSON(http.StatusOK, gin.H{})
}

// TestNewSSHKeyResource tests the SSHKeyResource can be properly provisioned.
func TestNewSSHKeyResource(t *testing.T) {
	t.Parallel()

	// Create a gin router
	router := gin.Default()

	lambdalabs.RegisterHandlers(router, nil)
	httptest.NewServer(router.Handler())
}
