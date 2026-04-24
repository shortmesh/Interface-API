import { useState, useEffect } from 'react'
import { Box, Typography, Button as MuiButton } from '@mui/material'
import { Add, Edit, Delete, Visibility, VisibilityOff, ContentCopy, Autorenew, Block } from '@mui/icons-material'
import {
  Table,
  Modal,
  Input,
  Tag,
  Space,
  message,
  Button,
  Alert as AntAlert,
} from 'antd'
import { apiCall, safeJsonParse, formatDate, maskString, copyToClipboard } from '../utils/api'

export default function Credentials() {
  const [credentials, setCredentials] = useState([])
  const [loading, setLoading] = useState(true)

  const [createOpen, setCreateOpen] = useState(false)
  const [newClientId, setNewClientId] = useState('')
  const [newDescription, setNewDescription] = useState('')
  const [creating, setCreating] = useState(false)

  const [secretModalOpen, setSecretModalOpen] = useState(false)
  const [revealedSecret, setRevealedSecret] = useState('')
  const [secretVisible, setSecretVisible] = useState(false)

  const [editOpen, setEditOpen] = useState(false)
  const [editTarget, setEditTarget] = useState(null)
  const [editDescription, setEditDescription] = useState('')
  const [regenerateSecret, setRegenerateSecret] = useState(false)
  const [deactivate, setDeactivate] = useState(false)
  const [updating, setUpdating] = useState(false)

  const [confirmRegenerateOpen, setConfirmRegenerateOpen] = useState(false)
  const [confirmDeactivateOpen, setConfirmDeactivateOpen] = useState(false)

  useEffect(() => {
    loadCredentials()
  }, [])

  const loadCredentials = async () => {
    setLoading(true)
    try {
      const response = await apiCall('/api/v1/admin/credentials')
      if (!response) return
      const data = await safeJsonParse(response)
      setCredentials(Array.isArray(data) ? data : [])
    } catch (error) {
      console.error('Error loading credentials:', error)
      message.error('Failed to load credentials')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!newClientId.trim()) {
      message.error('Client ID is required')
      return
    }
    setCreating(true)
    try {
      const response = await apiCall('/api/v1/admin/credentials', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ client_id: newClientId.trim(), description: newDescription.trim() }),
      })
      if (!response) return
      if (response.ok) {
        const data = await safeJsonParse(response)
        setCreateOpen(false)
        setNewClientId('')
        setNewDescription('')
        setRevealedSecret(data.client_secret)
        setSecretVisible(false)
        setSecretModalOpen(true)
        loadCredentials()
      } else {
        const err = await safeJsonParse(response)
        message.error(err.error || 'Failed to create credential')
      }
    } catch (error) {
      console.error('Error creating credential:', error)
      message.error('Failed to create credential')
    } finally {
      setCreating(false)
    }
  }

  const handleOpenEdit = (record) => {
    setEditTarget(record)
    setEditDescription(record.description || '')
    setRegenerateSecret(false)
    setDeactivate(false)
    setEditOpen(true)
  }

  const handleConfirmRegenerate = () => {
    setRegenerateSecret(true)
    setConfirmRegenerateOpen(false)
  }

  const handleConfirmDeactivate = () => {
    setDeactivate(true)
    setConfirmDeactivateOpen(false)
  }

  const handleUpdate = async () => {
    if (!editTarget) return
    setUpdating(true)
    try {
      const body = {}
      if (editDescription !== editTarget.description) {
        body.description = editDescription
      }
      if (regenerateSecret) {
        body.regenerate_secret = true
      }
      if (deactivate) {
        body.deactivate = true
      }

      const response = await apiCall(`/api/v1/admin/credentials/${editTarget.client_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!response) return
      if (response.ok) {
        const data = await safeJsonParse(response)
        setEditOpen(false)
        setEditTarget(null)
        loadCredentials()
        message.success('Credential updated successfully')
        if (data.client_secret) {
          setRevealedSecret(data.client_secret)
          setSecretVisible(false)
          setSecretModalOpen(true)
        }
      } else {
        const err = await safeJsonParse(response)
        message.error(err.error || 'Failed to update credential')
      }
    } catch (error) {
      console.error('Error updating credential:', error)
      message.error('Failed to update credential')
    } finally {
      setUpdating(false)
    }
  }

  const handleDelete = (clientId) => {
    Modal.confirm({
      title: 'Delete Credential',
      content: `Are you sure you want to permanently delete "${clientId}"? This cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          const response = await apiCall(`/api/v1/admin/credentials/${clientId}`, {
            method: 'DELETE',
          })
          if (!response) return
          if (response.ok) {
            loadCredentials()
            message.success('Credential deleted successfully')
          } else {
            const err = await safeJsonParse(response)
            message.error(err.error || 'Failed to delete credential')
          }
        } catch (error) {
          console.error('Error deleting credential:', error)
          message.error('Failed to delete credential')
        }
      },
    })
  }

  const handleCopy = async (text) => {
    try {
      await copyToClipboard(text)
      message.success('Copied to clipboard')
    } catch {
      message.error('Failed to copy')
    }
  }

  const columns = [
    {
      title: 'Client ID',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (text) => (
        <code style={{ fontFamily: '"Google Sans Mono", monospace' }}>{text}</code>
      ),
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role) => (
        <Tag color={role === 'super_admin' ? 'gold' : 'blue'}>
          {role === 'super_admin' ? 'Super Admin' : 'User'}
        </Tag>
      ),
    },
    {
      title: 'Scopes',
      dataIndex: 'scopes',
      key: 'scopes',
      render: (scopes) =>
        Array.isArray(scopes) && scopes.length > 0 ? (
          <Space size={[4, 4]} wrap>
            {scopes.map((s) => (
              <Tag key={s} style={{ fontFamily: '"Google Sans Mono", monospace', fontSize: 11 }}>
                {s}
              </Tag>
            ))}
          </Space>
        ) : (
          '-'
        ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      render: (text) => text || <Typography variant="caption" color="text.secondary">-</Typography>,
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      render: (active) => (
        <Tag color={active ? 'success' : 'error'}>{active ? 'Active' : 'Inactive'}</Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => formatDate(date),
    },
    {
      title: 'Actions',
      key: 'actions',
      align: 'center',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            icon={<Edit style={{ fontSize: 18 }} />}
            onClick={() => handleOpenEdit(record)}
          />
          <Button
            type="text"
            danger
            icon={<Delete style={{ fontSize: 18 }} />}
            onClick={() => handleDelete(record.client_id)}
          />
        </Space>
      ),
    },
  ]

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Box>
          <Typography variant="h5" gutterBottom>
            API Credentials
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage API credentials (client ID / client secret pairs) used to authenticate API requests.
          </Typography>
        </Box>
        <MuiButton
          variant="contained"
          startIcon={<Add />}
          onClick={() => {
            setNewClientId('')
            setNewDescription('')
            setCreateOpen(true)
          }}
        >
          New Credential
        </MuiButton>
      </Box>

      <Table
        dataSource={credentials}
        columns={columns}
        rowKey="client_id"
        loading={loading}
        pagination={{ pageSize: 10 }}
        scroll={{ x: true }}
      />

      <Modal
        title="Create Credential"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="Create"
        confirmLoading={creating}
        destroyOnClose
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          <Box>
            <Typography variant="body2" gutterBottom>
              Client ID <span style={{ color: 'red' }}>*</span>
            </Typography>
            <Input
              placeholder="e.g. my-app"
              value={newClientId}
              onChange={(e) => setNewClientId(e.target.value)}
              onPressEnter={handleCreate}
              autoFocus
            />
          </Box>
          <Box>
            <Typography variant="body2" gutterBottom>
              Description
            </Typography>
            <Input
              placeholder="Optional description"
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
            />
          </Box>
        </Box>
      </Modal>

      <Modal
        title={`Edit — ${editTarget?.client_id}`}
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleUpdate}
        okText="Save"
        confirmLoading={updating}
        destroyOnClose
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          <Box>
            <Typography variant="body2" gutterBottom>
              Description
            </Typography>
            <Input
              placeholder="Optional description"
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
            />
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Box>
              <Typography variant="body2">Regenerate client secret</Typography>
              <Typography variant="caption" color="text.secondary">
                Issue a new secret and invalidate the current one
              </Typography>
            </Box>
            {regenerateSecret ? (
              <AntAlert
                type="warning"
                showIcon
                banner
                message="Will regenerate on save"
                style={{ padding: '2px 8px', fontSize: 12 }}
              />
            ) : (
              <Button
                icon={<Autorenew style={{ fontSize: 16 }} />}
                onClick={() => setConfirmRegenerateOpen(true)}
              >
                Regenerate
              </Button>
            )}
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Box>
              <Typography variant="body2">Deactivate credential</Typography>
              <Typography variant="caption" color="text.secondary">
                Revoke API access immediately
              </Typography>
            </Box>
            {deactivate ? (
              <AntAlert
                type="error"
                showIcon
                banner
                message="Will deactivate on save"
                style={{ padding: '2px 8px', fontSize: 12 }}
              />
            ) : (
              <Button
                danger
                icon={<Block style={{ fontSize: 16 }} />}
                onClick={() => setConfirmDeactivateOpen(true)}
              >
                Deactivate
              </Button>
            )}
          </Box>
        </Box>
      </Modal>

      <Modal
        title="Regenerate Client Secret?"
        open={confirmRegenerateOpen}
        onCancel={() => setConfirmRegenerateOpen(false)}
        onOk={handleConfirmRegenerate}
        okText="Yes, regenerate"
        okButtonProps={{ danger: true }}
        cancelText="Cancel"
        zIndex={1100}
        width={420}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
          <AntAlert
            type="warning"
            showIcon
            message="This will immediately invalidate the current client secret."
            description="Any integrations using the old secret will stop working. The new secret will be shown once after saving — make sure to copy it."
          />
        </Box>
      </Modal>

      <Modal
        title="Deactivate Credential?"
        open={confirmDeactivateOpen}
        onCancel={() => setConfirmDeactivateOpen(false)}
        onOk={handleConfirmDeactivate}
        okText="Yes, deactivate"
        okButtonProps={{ danger: true }}
        cancelText="Cancel"
        zIndex={1100}
        width={420}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
          <AntAlert
            type="error"
            showIcon
            message={`Deactivating "${editTarget?.client_id}" will immediately revoke all API access.`}
            description="Any application or service using this credential will be denied access as soon as the change is saved. This action cannot be automatically reversed."
          />
        </Box>
      </Modal>

      <Modal
        title="Client Secret"
        open={secretModalOpen}
        onCancel={() => setSecretModalOpen(false)}
        footer={[
          <Button key="close" type="primary" onClick={() => setSecretModalOpen(false)}>
            Done
          </Button>,
        ]}
        destroyOnClose
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          <AntAlert
            type="warning"
            showIcon
            message="Save this secret now — it will not be shown again."
          />
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <code
              style={{
                flex: 1,
                wordBreak: 'break-all',
                fontFamily: '"Google Sans Mono", monospace',
                fontSize: 13,
                background: '#f5f5f5',
                padding: '6px 10px',
                borderRadius: 4,
              }}
            >
              {secretVisible ? revealedSecret : maskString(revealedSecret)}
            </code>
            <Button
              type="text"
              icon={
                secretVisible ? (
                  <VisibilityOff style={{ fontSize: 18 }} />
                ) : (
                  <Visibility style={{ fontSize: 18 }} />
                )
              }
              onClick={() => setSecretVisible((v) => !v)}
            />
            <Button
              type="text"
              icon={<ContentCopy style={{ fontSize: 18 }} />}
              onClick={() => handleCopy(revealedSecret)}
            />
          </Box>
        </Box>
      </Modal>
    </Box>
  )
}
