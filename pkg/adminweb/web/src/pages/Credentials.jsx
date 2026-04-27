import { useState, useEffect, useRef } from 'react'
import '../modalShake.css'
import { Box, Typography, Button as MuiButton } from '@mui/material'
import { Add, Edit, Delete, Visibility, VisibilityOff, ContentCopy, Autorenew, Close } from '@mui/icons-material'
import {
  Table,
  Modal,
  Input,
  Tag,
  Space,
  Switch,
  message,
  Button,
  Alert as AntAlert,
  App,
} from 'antd'
import { apiCall, safeJsonParse, formatDate, maskString, copyToClipboard } from '../utils/api'
import { hasScope } from '../utils/scopes'

export default function Credentials() {
  const { modal } = App.useApp()
  const canWrite = hasScope('credentials:write:*')
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
  const [desiredActive, setDesiredActive] = useState(true)
  const [updating, setUpdating] = useState(false)

  const [confirmRegenerateOpen, setConfirmRegenerateOpen] = useState(false)

  const [createError, setCreateError] = useState('')
  const [editError, setEditError] = useState('')

  const [createShake, setCreateShake] = useState(false)
  const [editShake, setEditShake] = useState(false)
  const [regenerateShake, setRegenerateShake] = useState(false)
  const [secretShake, setSecretShake] = useState(false)

  const closeByXRef = useRef(false)

  const triggerShake = (setter) => {
    setter(true)
    setTimeout(() => setter(false), 500)
  }

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
      setCreateError('Client ID is required')
      return
    }
    setCreateError('')
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
        setCreateError('')
        setRevealedSecret(data.client_secret)
        setSecretVisible(false)
        setSecretModalOpen(true)
        loadCredentials()
      } else {
        const err = await safeJsonParse(response)
        setCreateError(err.error || 'Failed to create credential')
      }
    } catch (error) {
      console.error('Error creating credential:', error)
      setCreateError('Failed to create credential')
    } finally {
      setCreating(false)
    }
  }

  const handleOpenEdit = (record) => {
    setEditTarget(record)
    setEditDescription(record.description || '')
    setRegenerateSecret(false)
    setDesiredActive(record.active)
    setEditError('')
    setEditOpen(true)
  }

  const handleConfirmRegenerate = () => {
    setRegenerateSecret(true)
    setConfirmRegenerateOpen(false)
  }

  const handleUpdate = async () => {
    if (!editTarget) return
    setEditError('')
    setUpdating(true)
    try {
      const body = {}
      if (editDescription !== editTarget.description) {
        body.description = editDescription
      }
      if (regenerateSecret) {
        body.regenerate_secret = true
      }
      if (desiredActive !== editTarget.active) {
        body.active = desiredActive
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
        setEditError('')
        loadCredentials()
        message.success('Credential updated successfully')
        if (data.client_secret) {
          setRevealedSecret(data.client_secret)
          setSecretVisible(false)
          setSecretModalOpen(true)
        }
      } else {
        const err = await safeJsonParse(response)
        setEditError(err.error || 'Failed to update credential')
      }
    } catch (error) {
      console.error('Error updating credential:', error)
      setEditError('Failed to update credential')
    } finally {
      setUpdating(false)
    }
  }

  const handleDelete = (clientId) => {
    modal.confirm({
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
            onClick={() =>
              canWrite
                ? handleOpenEdit(record)
                : message.info("You do not have permission to edit credentials. Contact admin.")
            }
            style={{ opacity: canWrite ? 1 : 0.45 }}
          />
          <Button
            type="text"
            danger={canWrite}
            icon={<Delete style={{ fontSize: 18 }} />}
            onClick={() =>
              canWrite
                ? handleDelete(record.client_id)
                : message.info("You do not have permission to delete credentials. Contact admin.")
            }
            style={{ opacity: canWrite ? 1 : 0.45 }}
          />
        </Space>
      ),
    },
  ]

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Box>
          <Typography variant="h6" gutterBottom>
            API Credentials
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage API credentials (client ID / client secret pairs) used to authenticate API requests.
          </Typography>
        </Box>
        <MuiButton
          variant="contained"
          startIcon={<Add />}
          onClick={() =>
            canWrite
              ? (setNewClientId(''), setNewDescription(''), setCreateError(''), setCreateOpen(true))
              : message.info("You do not have permission to create credentials. Contact admin.")
          }
          sx={{ opacity: canWrite ? 1 : 0.45 }}
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
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setCreateOpen(false); } else { triggerShake(setCreateShake); } }}
        maskClosable={true}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={createShake ? 'modal-shake' : ''}
        destroyOnHidden
        footer={[
          <Button key="cancel" onClick={() => setCreateOpen(false)} disabled={creating}>Cancel</Button>,
          <Button key="create" type="primary" onClick={handleCreate} loading={creating}>Create</Button>,
        ]}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          {createError && (
            <AntAlert type="error" showIcon message={createError} />
          )}
          <Box>
            <Typography variant="body2" gutterBottom>
              Client ID <span style={{ color: 'red' }}>*</span>
            </Typography>
            <Input
              placeholder="e.g. my-app"
              value={newClientId}
              onChange={(e) => { setNewClientId(e.target.value); setCreateError('') }}
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
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setEditOpen(false); } else { triggerShake(setEditShake); } }}
        maskClosable={true}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={editShake ? 'modal-shake' : ''}
        destroyOnHidden
        footer={[
          <Button key="cancel" onClick={() => setEditOpen(false)} disabled={updating}>Cancel</Button>,
          <Button key="save" type="primary" onClick={handleUpdate} loading={updating}>Save</Button>,
        ]}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          {editError && (
            <AntAlert type="error" showIcon message={editError} />
          )}
          <Box>
            <Typography variant="body2" gutterBottom>
              Description
            </Typography>
            <Input
              placeholder="Optional description"
              value={editDescription}
              onChange={(e) => { setEditDescription(e.target.value); setEditError('') }}
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
              <Typography variant="body2">
                {desiredActive ? 'Active' : 'Inactive'}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {desiredActive ? 'Toggle off to deactivate and revoke API access' : 'Toggle on to reactivate and restore API access'}
              </Typography>
            </Box>
            <Switch
              checked={desiredActive}
              onChange={setDesiredActive}
              checkedChildren="Active"
              unCheckedChildren="Inactive"
            />
          </Box>
        </Box>
      </Modal>

      <Modal
        title="Regenerate Client Secret?"
        open={confirmRegenerateOpen}
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setConfirmRegenerateOpen(false); } else { triggerShake(setRegenerateShake); } }}
        maskClosable={true}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={regenerateShake ? 'modal-shake' : ''}
        zIndex={1100}
        width={420}
        footer={[
          <Button key="cancel" onClick={() => setConfirmRegenerateOpen(false)}>Cancel</Button>,
          <Button key="regenerate" type="primary" danger onClick={handleConfirmRegenerate}>Yes, regenerate</Button>,
        ]}
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
        title="Client Secret"
        open={secretModalOpen}
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setSecretModalOpen(false); } else { triggerShake(setSecretShake); } }}
        closeIcon={<Close onClick={() => { closeByXRef.current = true; }} />}
        footer={[
          <Button key="close" type="primary" onClick={() => setSecretModalOpen(false)}>
            I have saved!
          </Button>,
        ]}
        maskClosable={true}
        wrapClassName={secretShake ? 'modal-shake' : ''}
        destroyOnHidden
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
          <AntAlert
            type="warning"
            showIcon
            message="Save this secret now — it will not be shown again."
          />
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              background: '#1e1e1e',
              border: '1px solid #333',
              borderRadius: 1,
              px: 1.5,
              py: 0.75,
              gap: 1,
            }}
          >
            <code
              style={{
                flex: 1,
                wordBreak: 'break-all',
                fontFamily: '"Google Sans Mono", monospace',
                fontSize: 13,
                color: '#e1e1e1',
                background: 'transparent',
              }}
            >
              {secretVisible ? revealedSecret : maskString(revealedSecret)}
            </code>
            <Button
              type="text"
              size="small"
              style={{ color: '#aaa', flexShrink: 0 }}
              icon={
                secretVisible ? (
                  <VisibilityOff style={{ fontSize: 16 }} />
                ) : (
                  <Visibility style={{ fontSize: 16 }} />
                )
              }
              onClick={() => setSecretVisible((v) => !v)}
            />
            <Button
              type="text"
              size="small"
              style={{ color: '#aaa', flexShrink: 0 }}
              icon={<ContentCopy style={{ fontSize: 16 }} />}
              onClick={() => handleCopy(revealedSecret)}
            />
          </Box>
        </Box>
      </Modal>
    </Box>
  )
}
