import { useState, useEffect, useRef } from "react";
import { Link } from 'react-router-dom'
import { Box, Typography, Button as MuiButton } from "@mui/material";
import '../modalShake.css'
import { Add, Edit, Delete, Close } from "@mui/icons-material";
import { Table, Modal, Input, Switch, Tag, Space, message, Spin, Button, App } from 'antd';
import { apiCall, safeJsonParse, formatDate } from "../utils/api";
import { hasScope } from "../utils/scopes";

export default function Webhooks() {
  const { modal } = App.useApp();
  const canWrite = hasScope("webhooks:write:*");
  const [webhooks, setWebhooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [matrixTokenDialogOpen, setMatrixTokenDialogOpen] = useState(false);
  const [webhookUrl, setWebhookUrl] = useState("");
  const [editWebhook, setEditWebhook] = useState(null);
  const [editUrl, setEditUrl] = useState("");
  const [editActive, setEditActive] = useState(true);
  const [matrixToken, setMatrixToken] = useState("");
  const [matrixTokenError, setMatrixTokenError] = useState("");
  const [matrixTokenShake, setMatrixTokenShake] = useState(false);  const [addWebhookShake, setAddWebhookShake] = useState(false)
  const [editWebhookShake, setEditWebhookShake] = useState(false)
  const [addWebhookError, setAddWebhookError] = useState("")
  const [editWebhookError, setEditWebhookError] = useState("")

  const closeByXRef = useRef(false)

  const triggerShake = (setter) => {
    setter(true)
    setTimeout(() => setter(false), 500)
  }
  useEffect(() => {
    loadWebhooks();
  }, []);

  const loadWebhooks = async () => {
    try {
      const tokenStatusResponse = await apiCall(
        "/api/v1/admin/matrix-token-status",
      );
      if (!tokenStatusResponse) return;
      const tokenStatus = await safeJsonParse(tokenStatusResponse);

      if (!tokenStatus.has_matrix_token) {
        setMatrixTokenDialogOpen(true);
        setLoading(false);
        return;
      }

      const response = await apiCall("/api/v1/admin/webhooks");
      if (!response) return;

      if (!response.ok) {
        if (response.status === 403) {
          setMatrixTokenDialogOpen(true);
          return;
        }
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to load webhooks");
        setWebhooks([]);
        return;
      }

      const data = await safeJsonParse(response);
      if (data?.error === 'Invalid server response') {
        setMatrixTokenDialogOpen(true);
        return;
      }
      setWebhooks(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Error loading webhooks:", error);
      message.error("Failed to load webhooks");
    } finally {
      setLoading(false);
    }
  };

  const handleSetMatrixToken = async () => {
    if (!matrixToken.trim()) {
      setMatrixTokenError("Matrix token is required");
      return;
    }
    if (!matrixToken.startsWith("mt_")) {
      setMatrixTokenError("Token must start with mt_");
      return;
    }

    try {
      const response = await apiCall("/api/v1/admin/matrix-token", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token: matrixToken }),
      });

      if (!response) return;

      if (response.ok) {
        setMatrixTokenDialogOpen(false);
        setMatrixToken("");
        setMatrixTokenError("");
        message.success("Matrix token set successfully");
        loadWebhooks();
      } else {
        const error = await safeJsonParse(response);
        setMatrixTokenError(error.error || "Failed to set token");
      }
    } catch (error) {
      console.error("Error setting token:", error);
      setMatrixTokenError("Failed to set token");
    }
  };

  const handleAddWebhook = async () => {
    if (!webhookUrl.trim()) {
      setAddWebhookError("Please enter a webhook URL")
      return
    }

    try {
      const response = await apiCall("/api/v1/admin/webhooks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: webhookUrl }),
      })

      if (!response) return

      if (response.ok) {
        setAddDialogOpen(false)
        setWebhookUrl("")
        setAddWebhookError("")
        loadWebhooks()
        message.success("Webhook added successfully")
      } else {
        if (response.status === 403) {
          setAddDialogOpen(false)
          setMatrixTokenDialogOpen(true)
          return
        }
        const error = await safeJsonParse(response)
        setAddWebhookError(error.error || "Failed to add webhook")
      }
    } catch (error) {
      console.error("Error adding webhook:", error)
      setAddWebhookError("Failed to add webhook")
    }
  }

  const handleUpdateWebhook = async () => {
    if (!editUrl.trim()) {
      setEditWebhookError("Please enter a webhook URL")
      return
    }

    try {
      const response = await apiCall(
        `/api/v1/admin/webhooks/${editWebhook.id}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ url: editUrl, active: editActive }),
        },
      )

      if (!response) return

      if (response.ok) {
        setEditDialogOpen(false)
        setEditWebhook(null)
        setEditWebhookError("")
        loadWebhooks()
        message.success("Webhook updated successfully")
      } else {
        if (response.status === 403) {
          setEditDialogOpen(false)
          setMatrixTokenDialogOpen(true)
          return
        }
        const error = await safeJsonParse(response)
        setEditWebhookError(error.error || "Failed to update webhook")
      }
    } catch (error) {
      console.error("Error updating webhook:", error)
      setEditWebhookError("Failed to update webhook")
    }
  }

  const handleDeleteWebhook = async (id) => {
    modal.confirm({
      title: 'Delete Webhook',
      content: 'Are you sure you want to delete this webhook?',
      okText: 'Delete',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          const response = await apiCall(`/api/v1/admin/webhooks/${id}`, {
            method: "DELETE",
          });

          if (!response) return;

          if (response.ok) {
            loadWebhooks();
            message.success("Webhook deleted successfully");
          } else {
            if (response.status === 403) {
              setMatrixTokenDialogOpen(true);
              return;
            }
            const error = await safeJsonParse(response);
            message.error(error.error || "Failed to delete webhook");
          }
        } catch (error) {
          console.error("Error deleting webhook:", error);
          message.error("Failed to delete webhook");
        }
      }
    });
  };

  const handleOpenEdit = (webhook) => {
    setEditWebhook(webhook);
    setEditUrl(webhook.url);
    setEditActive(webhook.active);
    setEditWebhookError("");
    setEditDialogOpen(true);
  };

  const handleOpenAdd = async () => {
    const tokenStatusResponse = await apiCall(
      "/api/v1/admin/matrix-token-status",
    )
    if (!tokenStatusResponse) return
    const tokenStatus = await safeJsonParse(tokenStatusResponse)

    if (!tokenStatus.has_matrix_token) {
      setMatrixTokenDialogOpen(true)
      return
    }

    setWebhookUrl("")
    setAddWebhookError("")
    setAddDialogOpen(true)
  }

  const columns = [
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      width: '40%',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      width: '10%',
      render: (active) => (
        <Tag color={active ? 'success' : 'error'}>
          {active ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '20%',
      render: (date) => formatDate(date),
    },
    {
      title: 'Updated',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: '20%',
      render: (date) => formatDate(date),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '10%',
      align: 'center',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            icon={<Edit style={{ fontSize: 18 }} />}
            onClick={() =>
              canWrite
                ? handleOpenEdit(record)
                : message.info("You do not have permission to edit webhooks. Contact admin.")
            }
            style={{ opacity: canWrite ? 1 : 0.45 }}
          />
          <Button
            type="text"
            icon={<Delete style={{ fontSize: 18 }} />}
            onClick={() =>
              canWrite
                ? handleDeleteWebhook(record.id)
                : message.info("You do not have permission to delete webhooks. Contact admin.")
            }
            danger={canWrite}
            style={{ opacity: canWrite ? 1 : 0.45 }}
          />
        </Space>
      ),
    },
  ];

  return (
    <Box>
    <Box>
      <Box
        sx={{
          mb: 4,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
        }}
      >
        <Box>
          <Typography variant="h6" gutterBottom>
            Webhooks
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage webhook URLs to receive message notifications.
          </Typography>
        </Box>
        <MuiButton
          variant="contained"
          startIcon={<Add />}
          onClick={() =>
            canWrite
              ? handleOpenAdd()
              : message.info("You do not have permission to add webhooks. Contact admin.")
          }
          sx={{ opacity: canWrite ? 1 : 0.45 }}
        >
          Add Webhook
        </MuiButton>
      </Box>
      </Box>

      <Table
        columns={columns}
        dataSource={webhooks}
        loading={loading}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: hasScope('webhooks:read:*') ? 'No webhooks found' : 'You do not have access to view webhooks. Contact admin.' }}
      />

      <Modal
        title="Add Webhook"
        open={addDialogOpen}
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setAddDialogOpen(false); } else { triggerShake(setAddWebhookShake); } }}
        maskClosable={true}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={addWebhookShake ? 'modal-shake' : ''}
        footer={[
          <Button variant="text" key="cancel" onClick={() => setAddDialogOpen(false)}>Cancel</Button>,
          <Button variant="text" key="add" type="primary" onClick={handleAddWebhook}>Add Webhook</Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <Input
            placeholder="https://example.com/webhook"
            value={webhookUrl}
            onChange={(e) => { setWebhookUrl(e.target.value); setAddWebhookError('') }}
            onPressEnter={handleAddWebhook}
            status={addWebhookError ? 'error' : ''}
          />
          {addWebhookError && (
            <div style={{ marginTop: 4, color: '#ff4d4f', fontSize: 12 }}>{addWebhookError}</div>
          )}
          <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255, 255, 255, 0.6)' }}>
            Enter a valid URL to receive webhook notifications
          </div>
        </div>
      </Modal>

      <Modal
        title="Edit Webhook"
        open={editDialogOpen}
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setEditDialogOpen(false); } else { triggerShake(setEditWebhookShake); } }}
        maskClosable={true}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={editWebhookShake ? 'modal-shake' : ''}
        footer={[
          <Button key="cancel" onClick={() => setEditDialogOpen(false)}>Cancel</Button>,
          <Button key="save" type="primary" onClick={handleUpdateWebhook}>Save Changes</Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <Input
            placeholder="Webhook URL"
            value={editUrl}
            onChange={(e) => { setEditUrl(e.target.value); setEditWebhookError('') }}
            onPressEnter={handleUpdateWebhook}
            status={editWebhookError ? 'error' : ''}
          />
          {editWebhookError && (
            <div style={{ marginTop: 4, color: '#ff4d4f', fontSize: 12 }}>{editWebhookError}</div>
          )}
          <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 14 }}>{editActive ? 'Active' : 'Inactive'}</div>
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.45)' }}>
                {editActive ? 'Toggle off to deactivate webhook' : 'Toggle on to reactivate webhook'}
              </div>
            </div>
            <Switch
              checked={editActive}
              onChange={setEditActive}
              checkedChildren="Active"
              unCheckedChildren="Inactive"
            />
          </div>
        </div>
      </Modal>

      <Modal
        title="Set Matrix Token"
        open={matrixTokenDialogOpen}
        onCancel={() => {
          setMatrixTokenShake(true)
          setTimeout(() => setMatrixTokenShake(false), 500)
        }}
        closable={false}
        maskClosable={true}
        wrapClassName={matrixTokenShake ? 'modal-shake' : ''}
        footer={[
          <Button key="skip" type="text" onClick={() => setMatrixTokenDialogOpen(false)}>
            Skip
          </Button>,
          <Button key="set" type="text" onClick={handleSetMatrixToken} style={{ color: '#4357AD' }}>
            Set Token
          </Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 8, color: 'rgba(255, 255, 255, 0.6)' }}>
            Please set your Matrix token to manage webhooks.
          </div>
          <Input
            placeholder="Matrix Token"
            value={matrixToken}
            onChange={(e) => {
              setMatrixToken(e.target.value);
              setMatrixTokenError("");
            }}
            status={matrixTokenError ? 'error' : ''}
            onPressEnter={handleSetMatrixToken}
          />
          {matrixTokenError && (
            <div style={{ marginTop: 4, color: '#ff4d4f', fontSize: 12 }}>
              {matrixTokenError}
            </div>
          )}
          <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255, 255, 255, 0.6)' }}>
            Your token will be stored for this session and will be cleared when you log out.
            <br />
            Don't have a token?{" "}
            <Link
              to="/tokens"
              style={{ color: "#4357AD", textDecoration: "underline" }}
            >
              Create one here
            </Link>
          </div>
        </div>
      </Modal>
    </Box>
  );
}
