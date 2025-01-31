/*
* Copyright (c) 2008-2023, Hazelcast, Inc. All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License")
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package codec

import (
	proto "github.com/hazelcast/hazelcast-go-client"
	hztypes "github.com/hazelcast/hazelcast-go-client/types"
)

const (
	TopicRemoveMessageListenerCodecRequestMessageType  = int32(0x040300)
	TopicRemoveMessageListenerCodecResponseMessageType = int32(0x040301)

	TopicRemoveMessageListenerCodecRequestRegistrationIdOffset = proto.PartitionIDOffset + proto.IntSizeInBytes
	TopicRemoveMessageListenerCodecRequestInitialFrameSize     = TopicRemoveMessageListenerCodecRequestRegistrationIdOffset + proto.UuidSizeInBytes

	TopicRemoveMessageListenerResponseResponseOffset = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
)

// Stops receiving messages for the given message listener.If the given listener already removed, this method does nothing.

func EncodeTopicRemoveMessageListenerRequest(name string, registrationId hztypes.UUID) *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, TopicRemoveMessageListenerCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	EncodeUUID(initialFrame.Content, TopicRemoveMessageListenerCodecRequestRegistrationIdOffset, registrationId)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(TopicRemoveMessageListenerCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	EncodeString(clientMessage, name)

	return clientMessage
}

func DecodeTopicRemoveMessageListenerResponse(clientMessage *proto.ClientMessage) bool {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	return DecodeBoolean(initialFrame.Content, TopicRemoveMessageListenerResponseResponseOffset)
}
